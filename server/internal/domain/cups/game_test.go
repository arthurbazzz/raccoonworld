package cups_test

import (
	"errors"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/cups"
	"raccoons-world/server/internal/domain/room"
)

// fixedShuffler always starts the nut in slot 0 and swaps slots 0 and 1, so
// after an odd number of swaps the nut is in slot 1, after an even in slot 0.
type fixedShuffler struct{}

func (fixedShuffler) Start() int { return 0 }
func (fixedShuffler) Swaps(n int) []cups.Swap {
	out := make([]cups.Swap, n)
	for i := range out {
		out[i] = cups.Swap{0, 1}
	}
	return out
}

var t0 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func newGame() *cups.Game { return cups.NewGame(cups.DefaultTiming(), fixedShuffler{}) }

// toChoose advances from the lobby to the Choose phase of round 1.
func toChoose(t *testing.T, g *cups.Game) time.Time {
	t.Helper()
	tm := cups.DefaultTiming()
	now := t0.Add(tm.Lobby + tm.Show + time.Duration(tm.Rounds[0].Swaps)*tm.Rounds[0].SwapTime)
	g.Advance(now)
	if g.Phase() != cups.Choose {
		t.Fatalf("phase = %v, want Choose", g.Phase())
	}
	return now
}

func find[T cups.Event](events []cups.Event) (T, bool) {
	for _, e := range events {
		if v, ok := e.(T); ok {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func TestFinalSlot(t *testing.T) {
	tests := []struct {
		start int
		swaps []cups.Swap
		want  int
	}{
		{0, nil, 0},
		{0, []cups.Swap{{0, 1}}, 1},
		{0, []cups.Swap{{1, 2}}, 0},
		{2, []cups.Swap{{1, 2}, {0, 1}}, 0},
		{1, []cups.Swap{{2, 1}, {2, 0}, {0, 1}}, 1},
	}
	for _, tt := range tests {
		if got := cups.FinalSlot(tt.start, tt.swaps); got != tt.want {
			t.Errorf("FinalSlot(%d, %v) = %d, want %d", tt.start, tt.swaps, got, tt.want)
		}
	}
}

func TestRandomShufflerSwapsTwoDifferentSlots(t *testing.T) {
	var s cups.RandomShuffler
	for range 200 {
		if n := s.Start(); n < 0 || n >= cups.Slots {
			t.Fatalf("Start() = %d", n)
		}
		for _, sw := range s.Swaps(10) {
			if sw[0] == sw[1] || sw[0] < 0 || sw[1] < 0 || sw[0] >= cups.Slots || sw[1] >= cups.Slots {
				t.Fatalf("bad swap %v", sw)
			}
		}
	}
}

func TestRoundsGetHarder(t *testing.T) {
	r := cups.DefaultTiming().Rounds
	if len(r) != 5 || r[0].Swaps != 3 || r[4].Swaps != 10 {
		t.Fatalf("rounds = %+v, want 5 rounds from 3 to 10 swaps (Q4)", r)
	}
	for i := 1; i < len(r); i++ {
		if r[i].Swaps < r[i-1].Swaps || r[i].SwapTime >= r[i-1].SwapTime {
			t.Fatalf("round %d is not harder than round %d", i+1, i)
		}
	}
}

func TestScaled(t *testing.T) {
	s := cups.DefaultTiming().Scaled(10)
	if s.Lobby != time.Second || s.Pick != time.Second || s.Rounds[0].SwapTime != 60*time.Millisecond || s.Rounds[0].Swaps != 3 {
		t.Fatalf("Scaled(10) = %+v", s)
	}
}

func TestLobby(t *testing.T) {
	g := newGame()
	res, ev := g.Join("rocky", t0)
	st, _ := find[cups.StatusChanged](ev)
	if res != cups.JoinedLobby || g.Phase() != cups.Lobby || st.LobbyLeft != 10*time.Second {
		t.Fatalf("first join: %v %v %+v, want a 10 s lobby (RC4)", res, g.Phase(), st)
	}

	res, ev = g.Join("bandit", t0.Add(4*time.Second))
	st, _ = find[cups.StatusChanged](ev)
	if res != cups.JoinedLobby || len(st.Players) != 2 || st.LobbyLeft != 6*time.Second {
		t.Fatalf("second join: %v %+v, want both in, 6 s left (RC6)", res, st)
	}

	if res, _ := g.Join("rocky", t0.Add(5*time.Second)); res != cups.JoinedNothing {
		t.Fatalf("repeated join = %v, want nothing", res)
	}

	ev = g.Advance(t0.Add(10 * time.Second))
	r, ok := find[cups.RoundStarted](ev)
	if !ok || r.Round != 1 || r.Total != 5 || len(r.Swaps) != 3 || g.Phase() != cups.Show {
		t.Fatalf("after the lobby: %+v, phase %v; want round 1 of 5 with 3 swaps", r, g.Phase())
	}
}

func TestJoinDuringMatchQueues(t *testing.T) {
	g := newGame()
	g.Join("rocky", t0)
	now := toChoose(t, g)
	res, ev := g.Join("bandit", now)
	st, _ := find[cups.StatusChanged](ev)
	if res != cups.JoinedQueue || len(st.Queued) != 1 || st.Queued[0] != "bandit" || !g.Busy("bandit") {
		t.Fatalf("join during match: %v %+v, want queued (RC5)", res, st)
	}
}

func TestPickRules(t *testing.T) {
	g := newGame()
	g.Join("rocky", t0)
	g.Join("bandit", t0)

	// During the show/shuffle picks are ignored (RC9).
	if _, err := g.Pick("rocky", 1, 0, t0.Add(10*time.Second+time.Millisecond)); !errors.Is(err, cups.ErrNotChoosing) {
		t.Fatalf("pick while shuffling: %v, want ErrNotChoosing", err)
	}

	now := toChoose(t, g)
	for _, tt := range []struct {
		id    room.PlayerID
		round int
		slot  int
		want  error
	}{
		{"stranger", 1, 0, cups.ErrNotInMatch},
		{"rocky", 2, 0, cups.ErrWrongRound},
		{"rocky", 1, 3, cups.ErrInvalidSlot},
		{"rocky", 1, -1, cups.ErrInvalidSlot},
	} {
		if _, err := g.Pick(tt.id, tt.round, tt.slot, now); !errors.Is(err, tt.want) {
			t.Errorf("Pick(%s, round %d, slot %d) = %v, want %v", tt.id, tt.round, tt.slot, err, tt.want)
		}
	}

	ev, err := g.Pick("rocky", 1, 1, now)
	if p, ok := find[cups.Picked](ev); err != nil || !ok || p.ID != "rocky" {
		t.Fatalf("first pick: %v %v", ev, err)
	}
	if _, err := g.Pick("rocky", 1, 2, now); !errors.Is(err, cups.ErrAlreadyPicked) {
		t.Fatalf("second pick: %v, want ErrAlreadyPicked", err)
	}
	if g.Phase() != cups.Choose {
		t.Fatal("cups went up before everyone picked")
	}

	// Everyone picked: cups go up now, not after 10 s (RC11).
	ev, _ = g.Pick("bandit", 1, 0, now)
	rv, ok := find[cups.Revealed](ev)
	if !ok || g.Phase() != cups.Reveal || rv.Nut != 1 {
		t.Fatalf("after all picked: %+v %v, want reveal with the nut in slot 1", rv, g.Phase())
	}
	want := []cups.PickResult{
		{ID: "rocky", Slot: 1, Picked: true, Hit: true},
		{ID: "bandit", Slot: 0, Picked: true, Hit: false},
	}
	for i, p := range want {
		if rv.Picks[i] != p {
			t.Fatalf("picks = %+v, want %+v (the first pick stays, RC9)", rv.Picks, want)
		}
	}
}

func TestTimeoutIsAMiss(t *testing.T) {
	g := newGame()
	g.Join("rocky", t0)
	now := toChoose(t, g)
	ev := g.Advance(now.Add(10*time.Second - time.Millisecond))
	if len(ev) != 0 {
		t.Fatalf("revealed before 10 s: %v", ev)
	}
	rv, ok := find[cups.Revealed](g.Advance(now.Add(10 * time.Second)))
	if !ok || rv.Picks[0].Picked || rv.Picks[0].Hit {
		t.Fatalf("after 10 s: %+v, want a miss without a pick (RC11)", rv)
	}
}

func TestFullMatchAndScoreboard(t *testing.T) {
	g := newGame()
	g.Join("rocky", t0)
	g.Join("bandit", t0)
	g.Join("ringo", t0)
	tm := cups.DefaultTiming()
	now := t0.Add(tm.Lobby)
	g.Advance(now)

	// rocky hits every round, bandit and ringo only odd rounds.
	var fin cups.Finished
	for round := 1; round <= 5; round++ {
		spec := tm.Rounds[round-1]
		now = now.Add(tm.Show + time.Duration(spec.Swaps)*spec.SwapTime)
		g.Advance(now)
		nut := spec.Swaps % 2 // fixedShuffler: odd swaps leave the nut in slot 1
		other := nut
		if round%2 == 0 {
			other = 2
		}
		mustPick(t, g, "rocky", round, nut, now)
		mustPick(t, g, "bandit", round, other, now)
		mustPick(t, g, "ringo", round, other, now)
		now = now.Add(tm.Reveal)
		if f, ok := find[cups.Finished](g.Advance(now)); ok {
			fin = f
		}
	}
	want := []cups.Score{{"rocky", 5, 1}, {"bandit", 3, 2}, {"ringo", 3, 2}}
	if g.Phase() != cups.Results || len(fin.Scores) != 3 {
		t.Fatalf("phase %v, scores %+v", g.Phase(), fin.Scores)
	}
	for i := range want {
		if fin.Scores[i] != want[i] {
			t.Fatalf("scores = %+v, want %+v (ties share a place, RC12)", fin.Scores, want)
		}
	}

	st, _ := find[cups.StatusChanged](g.Advance(now.Add(tm.Results)))
	if g.Phase() != cups.Idle || len(st.Players) != 0 || g.Busy("rocky") {
		t.Fatalf("after results: %v %+v, want the booth free", g.Phase(), st)
	}
}

func TestQueuedPlayersGetTheNextLobby(t *testing.T) {
	g := newGame()
	g.Join("rocky", t0)
	now := toChoose(t, g)
	g.Join("bandit", now)
	g.Leave("rocky", now)
	if g.Phase() != cups.Lobby || !g.Busy("bandit") || g.Busy("rocky") {
		t.Fatalf("phase %v, want the next lobby with bandit (RC5)", g.Phase())
	}
}

func TestLeave(t *testing.T) {
	g := newGame()
	g.Join("rocky", t0)
	g.Join("bandit", t0)
	now := toChoose(t, g)

	mustPick(t, g, "bandit", 1, 0, now)
	ok, ev := g.Leave("rocky", now)
	if !ok || g.Busy("rocky") {
		t.Fatal("rocky still in the match")
	}
	// bandit already picked: with rocky gone, cups go up (RC15).
	if _, revealed := find[cups.Revealed](ev); !revealed {
		t.Fatalf("events %v, want a reveal", ev)
	}
	if g.Phase() != cups.Reveal {
		t.Fatalf("phase %v, the match goes on for bandit (RC13)", g.Phase())
	}

	if ok, _ := g.Leave("stranger", now); ok {
		t.Fatal("stranger left a booth they never joined")
	}
	g.Leave("bandit", now)
	if g.Phase() != cups.Idle {
		t.Fatalf("phase %v, want Idle when everyone left (RC14)", g.Phase())
	}
}

func TestCurrentRound(t *testing.T) {
	g := newGame()
	if _, _, ok := g.CurrentRound(t0); ok {
		t.Fatal("idle booth has a round")
	}
	g.Join("rocky", t0)
	g.Advance(t0.Add(10 * time.Second))
	r, elapsed, ok := g.CurrentRound(t0.Add(11 * time.Second))
	if !ok || r.Round != 1 || elapsed != time.Second {
		t.Fatalf("CurrentRound = %+v %v %v", r, elapsed, ok)
	}
}

func mustPick(t *testing.T, g *cups.Game, id room.PlayerID, round, slot int, now time.Time) {
	t.Helper()
	if _, err := g.Pick(id, round, slot, now); err != nil {
		t.Fatalf("Pick(%s, round %d, slot %d): %v", id, round, slot, err)
	}
}

func TestPlayAgainFromTheScoreboard(t *testing.T) {
	g := cups.NewGame(cups.DefaultTiming(), fixedShuffler{})
	g.Join("rocky", t0)
	g.Join("bandit", t0)
	now := t0
	for g.Phase() != cups.Results {
		now = now.Add(50 * time.Millisecond)
		g.Advance(now)
	}
	if res, _ := g.Join("rocky", now); res != cups.JoinedQueue {
		t.Fatalf("play again = %v, want queued for the next match (RC12)", res)
	}
	if res, _ := g.Join("rocky", now); res != cups.JoinedNothing {
		t.Fatalf("play again twice = %v, want nothing", res)
	}
	st, _ := find[cups.StatusChanged](g.Advance(now.Add(cups.DefaultTiming().Results)))
	if g.Phase() != cups.Lobby || len(st.Players) != 1 || st.Players[0] != "rocky" || g.Busy("bandit") {
		t.Fatalf("after results: %v %+v, want a lobby with rocky only", g.Phase(), st)
	}
}
