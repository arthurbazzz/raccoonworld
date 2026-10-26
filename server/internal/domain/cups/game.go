package cups

import (
	"errors"
	"slices"
	"time"

	"raccoons-world/server/internal/domain/room"
)

// Phase is where a match is.
type Phase int

// The phases of the booth, in order (plan, section 3.1).
const (
	Idle    Phase = iota // nobody playing, nobody waiting
	Lobby                // countdown before the match (RC4)
	Show                 // the nut is shown under a cup
	Shuffle              // the cups swap places
	Choose               // players pick a cup (RC9)
	Reveal               // cups are up, showing the nut
	Results              // final scoreboard (RC12)
)

// Playing reports whether a round is in progress (Show through Reveal).
func (p Phase) Playing() bool { return p >= Show && p <= Reveal }

// Errors returned by Pick. The use case ignores them: a late or repeated
// pick is not a protocol error (RC9, RC18).
var (
	ErrNotInMatch    = errors.New("player is not in the match")
	ErrNotChoosing   = errors.New("not the time to pick")
	ErrWrongRound    = errors.New("pick is for another round")
	ErrAlreadyPicked = errors.New("player already picked this round")
	ErrInvalidSlot   = errors.New("invalid slot")
)

type entry struct {
	id     room.PlayerID
	hits   int
	pick   int // -1 until the player picks in the current round
	picked bool
}

// Game is the booth: at most one match at a time, with many players.
type Game struct {
	timing   Timing
	shuffler Shuffler

	phase   Phase
	ends    time.Time // when the current phase ends (not used in Idle)
	started time.Time // when the current round started

	players []*entry        // in the lobby or the match, in join order
	queued  []room.PlayerID // waiting for the next match (RC5)

	round int // 1-based; 0 before the first round
	nut   int // slot the nut started under this round
	swaps []Swap
}

// NewGame returns an idle booth.
func NewGame(t Timing, s Shuffler) *Game {
	return &Game{timing: t, shuffler: s}
}

// Phase returns the current phase.
func (g *Game) Phase() Phase { return g.phase }

// Busy reports whether id is in the lobby, the match or the queue: their
// raccoon stays put (RC21).
func (g *Game) Busy(id room.PlayerID) bool {
	return g.index(id) >= 0 || slices.Contains(g.queued, id)
}

// JoinResult says where Join put the player.
type JoinResult int

// Join outcomes.
const (
	JoinedNothing JoinResult = iota // already in, nothing changed
	JoinedLobby                     // in the lobby; the match starts with them
	JoinedQueue                     // a match is on; they play the next one (RC5)
)

// Join adds id to the booth (RC4, RC5). The first player opens the lobby.
func (g *Game) Join(id room.PlayerID, now time.Time) (JoinResult, []Event) {
	events := g.Advance(now)
	if g.phase == Results && g.index(id) >= 0 && !slices.Contains(g.queued, id) {
		// "Jogar de novo" on the scoreboard: play the next match (RC12).
		g.queued = append(g.queued, id)
		return JoinedQueue, append(events, g.status(now))
	}
	if g.Busy(id) {
		return JoinedNothing, events
	}
	switch g.phase {
	case Idle:
		g.phase = Lobby
		g.ends = now.Add(g.timing.Lobby)
		g.players = []*entry{{id: id, pick: -1}}
	case Lobby:
		g.players = append(g.players, &entry{id: id, pick: -1})
	default:
		g.queued = append(g.queued, id)
		return JoinedQueue, append(events, g.status(now))
	}
	return JoinedLobby, append(events, g.status(now))
}

// Pick records id's cup for round (RC9, RC10, RC18). When everyone in the
// match has picked, the cups go up right away (RC11).
func (g *Game) Pick(id room.PlayerID, round, slot int, now time.Time) ([]Event, error) {
	events := g.Advance(now)
	i := g.index(id)
	switch {
	case i < 0 || !(g.phase.Playing()):
		return events, ErrNotInMatch
	case round != g.round:
		return events, ErrWrongRound
	case g.phase != Choose:
		return events, ErrNotChoosing
	case slot < 0 || slot >= Slots:
		return events, ErrInvalidSlot
	case g.players[i].picked:
		return events, ErrAlreadyPicked
	}
	g.players[i].pick, g.players[i].picked = slot, true
	events = append(events, Picked{ID: id})
	if g.allPicked() {
		events = append(events, g.reveal(now)...)
	}
	return events, nil
}

// Leave takes id out of the booth (RC13, RC15). It reports whether they were
// there. With nobody left in the match the booth frees up (RC14), or opens
// the lobby for whoever is queued.
func (g *Game) Leave(id room.PlayerID, now time.Time) (bool, []Event) {
	events := g.Advance(now)
	if i := g.index(id); i >= 0 {
		g.players = slices.Delete(g.players, i, i+1)
	} else if j := slices.Index(g.queued, id); j >= 0 {
		g.queued = slices.Delete(g.queued, j, j+1)
	} else {
		return false, events
	}

	switch {
	case len(g.players) == 0 && g.phase != Idle:
		g.next(now)
	case g.phase == Choose && g.allPicked():
		events = append(events, g.reveal(now)...)
	}
	return true, append(events, g.status(now))
}

// Advance moves through every phase whose time is up and returns what
// happened. Each phase starts when the previous one was due, not when
// Advance runs, so the 50 ms tick does not stretch the match.
func (g *Game) Advance(now time.Time) []Event {
	var events []Event
	for g.phase != Idle && !now.Before(g.ends) {
		at := g.ends
		switch g.phase {
		case Lobby:
			events = append(events, g.startRound(1, at), g.status(at))
		case Show:
			g.phase, g.ends = Shuffle, at.Add(g.shuffleTime())
		case Shuffle:
			g.phase, g.ends = Choose, at.Add(g.timing.Pick)
		case Choose:
			events = append(events, g.reveal(at)...)
		case Reveal:
			if g.round < len(g.timing.Rounds) {
				events = append(events, g.startRound(g.round+1, at))
			} else {
				events = append(events, g.finish(at))
			}
		case Results:
			g.next(at)
			events = append(events, g.status(at))
		}
	}
	return events
}

// Status is who is at the booth and what it is doing (RC6, RC22).
func (g *Game) Status(now time.Time) StatusChanged {
	return g.status(now)
}

// CurrentRound returns the round in progress and how long ago it started,
// for players who arrive in the middle of it.
func (g *Game) CurrentRound(now time.Time) (RoundStarted, time.Duration, bool) {
	if !g.phase.Playing() {
		return RoundStarted{}, 0, false
	}
	return g.roundEvent(), now.Sub(g.started), true
}

func (g *Game) startRound(n int, at time.Time) Event {
	g.round, g.started = n, at
	g.nut = g.shuffler.Start()
	g.swaps = g.shuffler.Swaps(g.timing.Rounds[n-1].Swaps)
	for _, p := range g.players {
		p.pick, p.picked = -1, false
	}
	g.phase, g.ends = Show, at.Add(g.timing.Show)
	return g.roundEvent()
}

func (g *Game) roundEvent() RoundStarted {
	spec := g.timing.Rounds[g.round-1]
	return RoundStarted{
		Round:    g.round,
		Total:    len(g.timing.Rounds),
		Nut:      g.nut,
		Swaps:    slices.Clone(g.swaps),
		Show:     g.timing.Show,
		SwapTime: spec.SwapTime,
		Pick:     g.timing.Pick,
	}
}

func (g *Game) shuffleTime() time.Duration {
	spec := g.timing.Rounds[g.round-1]
	return time.Duration(len(g.swaps)) * spec.SwapTime
}

func (g *Game) reveal(at time.Time) []Event {
	nut := FinalSlot(g.nut, g.swaps)
	ev := Revealed{Round: g.round, Nut: nut, Picks: make([]PickResult, len(g.players))}
	for i, p := range g.players {
		hit := p.picked && p.pick == nut // no pick in time is a miss (RC11)
		if hit {
			p.hits++
		}
		ev.Picks[i] = PickResult{ID: p.id, Slot: p.pick, Picked: p.picked, Hit: hit}
	}
	g.phase, g.ends = Reveal, at.Add(g.timing.Reveal)
	return []Event{ev}
}

func (g *Game) finish(at time.Time) Event {
	scores := make([]Score, len(g.players))
	for i, p := range g.players {
		scores[i] = Score{ID: p.id, Hits: p.hits}
	}
	slices.SortStableFunc(scores, func(a, b Score) int { return b.Hits - a.Hits })
	for i := range scores {
		// Ties share a place (RC12): 1 + how many did better.
		scores[i].Place = 1 + slices.IndexFunc(scores, func(s Score) bool { return s.Hits == scores[i].Hits })
	}
	g.phase, g.ends = Results, at.Add(g.timing.Results)
	return Finished{Scores: scores}
}

// next frees the booth or, if someone is queued, opens the next lobby.
func (g *Game) next(at time.Time) {
	g.round, g.swaps = 0, nil
	g.players = g.players[:0]
	if len(g.queued) == 0 {
		g.phase = Idle
		return
	}
	for _, id := range g.queued {
		g.players = append(g.players, &entry{id: id, pick: -1})
	}
	g.queued = nil
	g.phase, g.ends = Lobby, at.Add(g.timing.Lobby)
}

func (g *Game) allPicked() bool {
	for _, p := range g.players {
		if !p.picked {
			return false
		}
	}
	return len(g.players) > 0
}

func (g *Game) status(now time.Time) StatusChanged {
	s := StatusChanged{Phase: g.phase, Players: make([]room.PlayerID, len(g.players)), Queued: slices.Clone(g.queued)}
	for i, p := range g.players {
		s.Players[i] = p.id
	}
	if s.Queued == nil {
		s.Queued = []room.PlayerID{}
	}
	if g.phase == Lobby {
		s.LobbyLeft = max(0, g.ends.Sub(now))
	}
	return s
}

func (g *Game) index(id room.PlayerID) int {
	return slices.IndexFunc(g.players, func(p *entry) bool { return p.id == id })
}
