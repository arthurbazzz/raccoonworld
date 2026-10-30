package rummy

import (
	"errors"
	"slices"
	"time"

	"raccoons-world/server/internal/domain/room"
)

// Errors returned by the turn actions. They answer a single player and do
// not change the match.
var (
	ErrNotInMatch  = errors.New("player is not in the match")
	ErrNotYourTurn = errors.New("not this player's turn")
	ErrStaleTurn   = errors.New("action is for another turn")
	ErrBadLayout   = errors.New("layout does not hold the same tiles")
	ErrMatchOver   = errors.New("match is over")
)

// Reasons a turn is refused, besides the meld errors (sent to the client).
var (
	ErrNothingPlaced = errors.New("nothing_placed")
	ErrOpeningLow    = errors.New("opening_too_low")
	ErrOpeningTable  = errors.New("opening_used_table")
)

// Result is how a turn ended (plan, section 3.5).
type Result string

// The turn results.
const (
	ResultOK      Result = "ok"      // the play was accepted
	ResultDrew    Result = "drew"    // chose to draw, or played nothing
	ResultTimeout Result = "timeout" // ran out of time (RN15)
)

// Player is someone sitting at the match.
type Player struct {
	ID       room.PlayerID
	Nickname string
}

// Layout is a full arrangement of the table and one hand: what the client
// sends while dragging tiles (plan, section 3.1). Pos has one entry per
// meld.
type Layout struct {
	Melds [][]TileID
	Pos   []Pos
	Hand  []TileID
}

// The table is a grid where each meld has a position (RN32, plan 9.3).
const (
	TableCols = 48
	TableRows = 28
)

// Pos is where a meld lies on the table, in grid cells.
type Pos struct{ X, Y int }

// AutoPos lays n melds out in rows of three, for a layout sent without
// positions (tests and older clients).
func AutoPos(n int) []Pos {
	pos := make([]Pos, n)
	for i := range pos {
		pos[i] = Pos{X: i % 3 * (TableCols / 3), Y: i / 3 * 3 % TableRows}
	}
	return pos
}

func (l Layout) clone() Layout {
	melds := make([][]TileID, len(l.Melds))
	for i, m := range l.Melds {
		melds[i] = slices.Clone(m)
	}
	return Layout{Melds: melds, Pos: slices.Clone(l.Pos), Hand: slices.Clone(l.Hand)}
}

type seat struct {
	Player
	hand        []TileID
	opened      bool
	absent      bool
	absentSince time.Time
}

// Match is one game at the table, from the deal to the winner.
type Match struct {
	timing   Timing
	mode     Mode
	shuffler Shuffler

	seats    []*seat // in turn order
	current  int     // index in seats of whose turn it is
	turn     int     // increases every turn, so late actions are ignored
	turnEnds time.Time
	pool     []TileID
	table    [][]TileID // confirmed melds
	tablePos []Pos      // where each confirmed meld lies
	draft    Layout     // the current player's arrangement this turn
	passes   int        // turns in a row ended without a tile, with an empty pool

	over   bool
	result MatchOver
}

// NewMatch shuffles, draws the turn order and deals 14 tiles to each player
// (RN8). players must have between MinPlayers and MaxPlayers entries.
func NewMatch(mode Mode, players []Player, s Shuffler, t Timing, now time.Time) (*Match, []Event) {
	m := &Match{timing: t, mode: mode, shuffler: s, pool: NewDeck()}
	s.Shuffle(m.pool)

	// The turn order is drawn with the same shuffler, over seat numbers.
	order := make([]TileID, len(players))
	for i := range order {
		order[i] = TileID(i)
	}
	s.Shuffle(order)
	for _, i := range order {
		hand := slices.Clone(m.pool[:HandSize])
		m.pool = m.pool[HandSize:]
		m.seats = append(m.seats, &seat{Player: players[i], hand: hand})
	}
	return m, []Event{m.startTurn(now)}
}

func (m *Match) startTurn(now time.Time) TurnStarted {
	m.turn++
	m.turnEnds = now.Add(m.timing.Turn)
	m.draft = m.startLayout()
	return TurnStarted{Turn: m.turn, Player: m.seats[m.current].ID, Ends: m.turnEnds}
}

// startLayout is the table and the current hand as the turn started.
func (m *Match) startLayout() Layout {
	return Layout{Melds: m.table, Pos: m.tablePos, Hand: m.seats[m.current].hand}.clone()
}

func (m *Match) seatOf(id room.PlayerID) (int, bool) {
	for i, s := range m.seats {
		if s.ID == id {
			return i, true
		}
	}
	return 0, false
}

// checkTurn makes sure id may act on turn now.
func (m *Match) checkTurn(id room.PlayerID, turn int) error {
	if m.over {
		return ErrMatchOver
	}
	i, ok := m.seatOf(id)
	if !ok {
		return ErrNotInMatch
	}
	if i != m.current {
		return ErrNotYourTurn
	}
	if turn != m.turn {
		return ErrStaleTurn
	}
	return nil
}

// Move replaces the current player's draft (RN11, RN12). Melds may be
// invalid for now; only the tiles are checked: the layout must hold exactly
// the table and hand tiles of the start of the turn, and no table tile may
// go to the hand. Positions must be inside the table (RN32).
func (m *Match) Move(id room.PlayerID, turn int, l Layout) error {
	if err := m.checkTurn(id, turn); err != nil {
		return err
	}
	onTable := map[TileID]bool{}
	want := map[TileID]bool{}
	for _, meld := range m.table {
		for _, t := range meld {
			onTable[t] = true
			want[t] = true
		}
	}
	for _, t := range m.seats[m.current].hand {
		want[t] = true
	}
	seen := map[TileID]bool{}
	use := func(t TileID) bool {
		if !want[t] || seen[t] {
			return false
		}
		seen[t] = true
		return true
	}
	if l.Pos == nil {
		l.Pos = AutoPos(len(l.Melds))
	}
	if len(l.Pos) != len(l.Melds) {
		return ErrBadLayout
	}
	melds := make([][]TileID, 0, len(l.Melds))
	pos := make([]Pos, 0, len(l.Melds))
	for i, meld := range l.Melds {
		if len(meld) == 0 {
			continue
		}
		if p := l.Pos[i]; p.X < 0 || p.Y < 0 || p.X >= TableCols || p.Y >= TableRows {
			return ErrBadLayout
		}
		pos = append(pos, l.Pos[i])
		for _, t := range meld {
			if !use(t) {
				return ErrBadLayout
			}
		}
		melds = append(melds, meld)
	}
	for _, t := range l.Hand {
		if onTable[t] || !use(t) {
			return ErrBadLayout
		}
	}
	if len(seen) != len(want) {
		return ErrBadLayout
	}
	m.draft = Layout{Melds: melds, Pos: pos, Hand: l.Hand}.clone()
	return nil
}

// Undo puts the draft back as it was at the start of the turn (RN12).
func (m *Match) Undo(id room.PlayerID, turn int) error {
	if err := m.checkTurn(id, turn); err != nil {
		return err
	}
	m.draft = m.startLayout()
	return nil
}

// EndTurn checks the draft (RN2, RN11, RN13, RN14). A valid play becomes the
// table. A refused one stays as the player left it and the turn goes on, so
// they can fix it while there is time (Q12).
func (m *Match) EndTurn(id room.PlayerID, turn int, now time.Time) ([]Event, error) {
	if err := m.checkTurn(id, turn); err != nil {
		return nil, err
	}
	s := m.seats[m.current]
	if len(m.draft.Hand) == len(s.hand) {
		return []Event{m.refuse(Issue{Meld: -1, Code: ErrNothingPlaced.Error()})}, nil
	}
	if issues := m.checkDraft(s); len(issues) > 0 {
		return []Event{m.refuse(issues...)}, nil
	}

	d := m.draft.clone()
	m.table, m.tablePos = d.Melds, d.Pos
	s.hand = slices.Clone(m.draft.Hand)
	s.opened = true
	m.passes = 0
	events := []Event{TurnEnded{Turn: m.turn, By: s.ID, Result: ResultOK}}
	if len(s.hand) == 0 {
		return append(events, m.finish(s.ID)), nil
	}
	return append(events, m.nextTurn(now)), nil
}

func (m *Match) refuse(issues ...Issue) Refused {
	return Refused{Turn: m.turn, Player: m.seats[m.current].ID, Issues: issues}
}

// checkDraft returns what is wrong with the draft: every invalid meld, or
// else what breaks the opening (RN12a, RN13).
func (m *Match) checkDraft(s *seat) []Issue {
	var issues []Issue
	for i, meld := range m.draft.Melds {
		if _, issue := Diagnose(meld); issue != nil {
			issue.Meld = i
			issues = append(issues, *issue)
		}
	}
	if len(issues) > 0 || m.mode != Classic || s.opened {
		return issues
	}
	// Opening (RN13): the table melds stay as they were, and the new ones,
	// made only of hand tiles, add up to at least 30 points.
	left := slices.Clone(m.table)
	points := 0
	inHand := map[TileID]bool{}
	for _, t := range s.hand {
		inHand[t] = true
	}
	for i, meld := range m.draft.Melds {
		if j := slices.IndexFunc(left, func(o []TileID) bool { return slices.Equal(o, meld) }); j >= 0 {
			left = slices.Delete(left, j, j+1)
			continue
		}
		var fromTable []TileID
		for _, t := range meld {
			if !inHand[t] {
				fromTable = append(fromTable, t)
			}
		}
		if len(fromTable) > 0 {
			issues = append(issues, Issue{Meld: i, Code: ErrOpeningTable.Error(), Tiles: fromTable})
		}
		points += Points(meld)
	}
	if len(issues) == 0 && len(left) > 0 {
		issues = append(issues, Issue{Meld: -1, Code: ErrOpeningTable.Error(), Tiles: slices.Concat(left...)})
	}
	if len(issues) == 0 && points < OpeningPoints {
		issues = append(issues, Issue{Meld: -1, Code: ErrOpeningLow.Error(), Params: map[string]int{
			"points": points, "need": OpeningPoints - points,
		}})
	}
	return issues
}

// Draw undoes the draft and draws a tile, ending the turn (RN10). With an
// empty pool the player passes (RN16).
func (m *Match) Draw(id room.PlayerID, turn int, now time.Time) ([]Event, error) {
	if err := m.checkTurn(id, turn); err != nil {
		return nil, err
	}
	return m.drawAndPass(ResultDrew, now), nil
}

// drawAndPass ends the current turn without a play: the draft is dropped
// and the player draws a tile if there is one.
func (m *Match) drawAndPass(r Result, now time.Time) []Event {
	s := m.seats[m.current]
	ev := TurnEnded{Turn: m.turn, By: s.ID, Result: r}
	if len(m.pool) > 0 {
		s.hand = append(s.hand, m.pool[0])
		m.pool = m.pool[1:]
		ev.Drew = true
	} else {
		m.passes++
	}
	if m.passes >= len(m.seats) {
		return []Event{ev, m.finish("")}
	}
	return []Event{ev, m.nextTurn(now)}
}

func (m *Match) nextTurn(now time.Time) TurnStarted {
	m.current = (m.current + 1) % len(m.seats)
	return m.startTurn(now)
}

// Advance applies the deadlines: a turn that ran out of time is undone and
// the player draws (RN15), and a player absent for too long leaves (RN21).
func (m *Match) Advance(now time.Time) []Event {
	var events []Event
	for _, s := range slices.Clone(m.seats) {
		if !m.over && s.absent && !now.Before(s.absentSince.Add(m.timing.Absence)) {
			events = append(events, m.Leave(s.ID, now)...)
		}
	}
	if !m.over && !now.Before(m.turnEnds) {
		events = append(events, m.drawAndPass(ResultTimeout, now)...)
	}
	return events
}

// Leave takes a player out (RN20): their tiles go back to the pool, which is
// shuffled again. If it was their turn, the draft is dropped and the turn
// passes on. A single player left wins (RN22).
func (m *Match) Leave(id room.PlayerID, now time.Time) []Event {
	i, ok := m.seatOf(id)
	if m.over || !ok {
		return nil
	}
	s := m.seats[i]
	m.pool = append(m.pool, s.hand...)
	m.shuffler.Shuffle(m.pool)
	m.seats = slices.Delete(m.seats, i, i+1)
	events := []Event{PlayerRemoved{Player: s.Player}}

	if len(m.seats) == 1 {
		return append(events, m.finish(m.seats[0].ID))
	}
	switch {
	case i < m.current:
		m.current--
	case i == m.current:
		// The next player slid into this index.
		m.current %= len(m.seats)
		events = append(events, m.startTurn(now))
	}
	return events
}

// SetAbsent marks a player whose connection dropped (RN21). The seat is kept
// for Timing.Absence, and the turn clock keeps running.
func (m *Match) SetAbsent(id room.PlayerID, now time.Time) {
	if i, ok := m.seatOf(id); ok && !m.seats[i].absent {
		m.seats[i].absent = true
		m.seats[i].absentSince = now
	}
}

// Reclaim gives an absent seat with this nickname to the player who came
// back with it, under a new ID (plan, section 3.2). It returns the old ID.
func (m *Match) Reclaim(nickname string, id room.PlayerID) (room.PlayerID, bool) {
	for _, s := range m.seats {
		if s.absent && s.Nickname == nickname {
			old := s.ID
			s.ID = id
			s.absent = false
			return old, true
		}
	}
	return "", false
}

// finish ends the match. With no winner (empty pool, RN18), the lowest hand
// wins; ties share the place.
func (m *Match) finish(winner room.PlayerID) MatchOver {
	scores := make([]Score, len(m.seats))
	for i, s := range m.seats {
		scores[i] = Score{Player: s.Player, Points: HandPoints(s.hand)}
	}
	slices.SortStableFunc(scores, func(a, b Score) int { return a.Points - b.Points })
	for i := range scores {
		scores[i].Place = i + 1
		if i > 0 && scores[i].Points == scores[i-1].Points {
			scores[i].Place = scores[i-1].Place
		}
	}
	if winner == "" && (len(scores) == 1 || scores[0].Points < scores[1].Points) {
		winner = scores[0].ID
	}
	m.over = true
	m.result = MatchOver{Winner: winner, Scores: scores}
	return m.result
}

// HandPoints is what a hand left at the end is worth: the numbers, and 30
// per joker (RN18).
func HandPoints(hand []TileID) int {
	sum := 0
	for _, id := range hand {
		if t := TileOf(id); t.Joker {
			sum += JokerPenalty
		} else {
			sum += t.Value
		}
	}
	return sum
}

// Mode returns the match mode.
func (m *Match) Mode() Mode { return m.mode }

// Over reports whether the match has ended, and how.
func (m *Match) Over() (MatchOver, bool) { return m.result, m.over }

// Has reports whether id has a seat in the match.
func (m *Match) Has(id room.PlayerID) bool {
	_, ok := m.seatOf(id)
	return ok
}

// Players returns who is seated, in turn order.
func (m *Match) Players() []Player {
	out := make([]Player, len(m.seats))
	for i, s := range m.seats {
		out[i] = s.Player
	}
	return out
}

// View is the match as one player may see it (RN9, RN24): their own hand
// only, and how many tiles everyone else has.
type View struct {
	Turn     int
	Current  room.PlayerID
	TurnEnds time.Time
	Mode     Mode
	Table    [][]TileID // confirmed table, or the viewer's draft on their turn
	Pos      []Pos      // where each meld of Table lies
	Hand     []TileID   // the viewer's hand, or their draft hand on their turn
	Counts   map[room.PlayerID]int
	Pool     int
	Opened   []room.PlayerID
	Absent   []room.PlayerID
}

// ViewFor returns what id sees.
func (m *Match) ViewFor(id room.PlayerID) View {
	cur := m.seats[m.current]
	v := View{
		Turn: m.turn, Current: cur.ID, TurnEnds: m.turnEnds, Mode: m.mode,
		Table: m.table, Pos: m.tablePos, Counts: map[room.PlayerID]int{}, Pool: len(m.pool),
	}
	for _, s := range m.seats {
		v.Counts[s.ID] = len(s.hand)
		if s.opened {
			v.Opened = append(v.Opened, s.ID)
		}
		if s.absent {
			v.Absent = append(v.Absent, s.ID)
		}
		if s.ID == id {
			v.Hand = s.hand
		}
	}
	if id == cur.ID && !m.over {
		v.Table, v.Pos, v.Hand = m.draft.Melds, m.draft.Pos, m.draft.Hand
	}
	c := Layout{Melds: v.Table, Pos: v.Pos, Hand: v.Hand}.clone()
	v.Table, v.Pos, v.Hand = c.Melds, c.Pos, c.Hand
	return v
}

// Draft returns the current player's arrangement, which the others watch
// live (plan, section 3.5): the melds and how many tiles are left in hand.
func (m *Match) Draft() (turn int, melds [][]TileID, pos []Pos, handCount int) {
	d := m.draft.clone()
	return m.turn, d.Melds, d.Pos, len(d.Hand)
}
