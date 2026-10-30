package rummy_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/domain/rummy"
)

var t0 = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

// stacked deals known hands: seat i gets hands[i] (padded to 14 with tiles
// nobody asked for), in seat order, and the pool keeps deck order. Shuffles
// of anything else (seat order, pool) leave things as they are.
type stacked struct{ deck []rummy.TileID }

func (s stacked) Shuffle(tiles []rummy.TileID) {
	if len(tiles) == rummy.DeckSize {
		copy(tiles, s.deck)
	}
}

// stackHands pads hands with tiles from the end of the deck, skipping the
// jokers so they stay in the pool.
func stackHands(t *testing.T, hands ...[]string) stacked {
	t.Helper()
	used := map[rummy.TileID]bool{}
	take := func(name string) rummy.TileID {
		for _, id := range rummy.NewDeck() {
			if !used[id] && rummy.TileOf(id).String() == name {
				used[id] = true
				return id
			}
		}
		t.Fatalf("no tile %q left", name)
		return 0
	}
	ids := make([][]rummy.TileID, len(hands))
	for i, hand := range hands {
		for _, name := range hand {
			ids[i] = append(ids[i], take(name))
		}
	}
	var deck []rummy.TileID
	next := rummy.TileID(2*rummy.Colors*rummy.MaxValue - 1)
	for _, hand := range ids {
		for len(hand) < rummy.HandSize {
			if !used[next] {
				used[next] = true
				hand = append(hand, next)
			}
			next--
		}
		deck = append(deck, hand...)
	}
	for _, id := range rummy.NewDeck() {
		if !used[id] {
			deck = append(deck, id)
		}
	}
	return stacked{deck: deck}
}

var (
	rocky  = rummy.Player{ID: "p1", Nickname: "rocky"}
	bandit = rummy.Player{ID: "p2", Nickname: "bandit"}
	mel    = rummy.Player{ID: "p3", Nickname: "mel"}
)

func newMatch(t *testing.T, mode rummy.Mode, hands ...[]string) *rummy.Match {
	t.Helper()
	players := []rummy.Player{rocky, bandit, mel}[:len(hands)]
	m, _ := rummy.NewMatch(mode, players, stackHands(t, hands...), rummy.DefaultTiming(), t0)
	return m
}

// find returns the IDs in hand with these faces, in this order.
func find(t *testing.T, hand []rummy.TileID, names ...string) []rummy.TileID {
	t.Helper()
	var out []rummy.TileID
	left := slices.Clone(hand)
	for _, n := range names {
		i := slices.IndexFunc(left, func(id rummy.TileID) bool { return rummy.TileOf(id).String() == n })
		if i < 0 {
			t.Fatalf("%q is not in the hand", n)
		}
		out = append(out, left[i])
		left = slices.Delete(left, i, i+1)
	}
	return out
}

func without(hand, ids []rummy.TileID) []rummy.TileID {
	return slices.DeleteFunc(slices.Clone(hand), func(id rummy.TileID) bool { return slices.Contains(ids, id) })
}

// play lays new melds from the current player's hand, on top of the table.
func play(t *testing.T, m *rummy.Match, id room.PlayerID, melds ...[]string) []rummy.Event {
	t.Helper()
	v := m.ViewFor(id)
	table := v.Table
	hand := v.Hand
	for _, names := range melds {
		ids := find(t, hand, names...)
		table = append(table, ids)
		hand = without(hand, ids)
	}
	if err := m.Move(id, v.Turn, rummy.Layout{Melds: table, Hand: hand}); err != nil {
		t.Fatalf("Move: %v", err)
	}
	ev, err := m.EndTurn(id, v.Turn, t0)
	if err != nil {
		t.Fatalf("EndTurn: %v", err)
	}
	return ev
}

// refused returns the codes of the Refused event in events.
func refused(t *testing.T, events []rummy.Event) []string {
	t.Helper()
	for _, e := range events {
		if r, ok := e.(rummy.Refused); ok {
			codes := make([]string, len(r.Issues))
			for i, is := range r.Issues {
				codes[i] = is.Code
			}
			return codes
		}
	}
	t.Fatalf("no Refused in %v", events)
	return nil
}

func ended(t *testing.T, events []rummy.Event) rummy.TurnEnded {
	t.Helper()
	for _, e := range events {
		if te, ok := e.(rummy.TurnEnded); ok {
			return te
		}
	}
	t.Fatalf("no TurnEnded in %v", events)
	return rummy.TurnEnded{}
}

func TestDeal(t *testing.T) {
	m := newMatch(t, rummy.Free, nil, nil)
	v := m.ViewFor(rocky.ID)
	if len(v.Hand) != 14 || v.Counts[bandit.ID] != 14 || v.Pool != 78 {
		t.Fatalf("hand %d, bandit %d, pool %d; want 14, 14, 78 (RN8)", len(v.Hand), v.Counts[bandit.ID], v.Pool)
	}
	if v.Current != rocky.ID || v.Turn != 1 {
		t.Fatalf("turn %d of %s, want 1 of rocky", v.Turn, v.Current)
	}
}

func TestViewHidesOtherHands(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{"red-1"}, []string{"blue-1"})
	for _, id := range m.ViewFor(bandit.ID).Hand {
		if rummy.TileOf(id).String() == "red-1" {
			t.Fatal("bandit sees rocky's tile (RN9)")
		}
	}
}

func TestMoveMustKeepTheSameTiles(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{"blue-3", "blue-4", "blue-5"}, nil)
	play(t, m, rocky.ID, []string{"blue-3", "blue-4", "blue-5"})
	v := m.ViewFor(bandit.ID)
	tests := []struct {
		name string
		l    rummy.Layout
	}{
		{"table tile to the hand", rummy.Layout{Hand: append(slices.Clone(v.Hand), v.Table[0]...)}},
		{"tile missing", rummy.Layout{Melds: v.Table, Hand: v.Hand[1:]}},
		{"tile invented", rummy.Layout{Melds: v.Table, Hand: append(slices.Clone(v.Hand), 105)}},
		{"tile twice", rummy.Layout{Melds: append(slices.Clone(v.Table), v.Table[0]), Hand: v.Hand}},
	}
	for _, tt := range tests {
		if err := m.Move(bandit.ID, v.Turn, tt.l); !errors.Is(err, rummy.ErrBadLayout) {
			t.Errorf("%s: Move = %v, want ErrBadLayout", tt.name, err)
		}
	}
	if err := m.Move(rocky.ID, v.Turn, rummy.Layout{}); !errors.Is(err, rummy.ErrNotYourTurn) {
		t.Fatalf("Move out of turn = %v", err)
	}
	if err := m.Move(bandit.ID, v.Turn-1, rummy.Layout{Melds: v.Table, Hand: v.Hand}); !errors.Is(err, rummy.ErrStaleTurn) {
		t.Fatalf("Move for an old turn = %v", err)
	}
}

func TestInvalidPlayKeepsTheTurn(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{"blue-7", "green-7", "red-7"}, nil)
	ev := play(t, m, rocky.ID, []string{"blue-7", "green-7"})
	if got := refused(t, ev); !slices.Equal(got, []string{"too_short"}) {
		t.Fatalf("Refused = %v, want too_short (RN14)", got)
	}
	v := m.ViewFor(rocky.ID)
	if v.Current != rocky.ID || len(v.Table) != 1 || v.Counts[rocky.ID] != 14 {
		t.Fatalf("current %s, draft %v, %d tiles; want rocky's turn with the draft kept and no draw (Q12)",
			v.Current, v.Table, v.Counts[rocky.ID])
	}

	// Fixing the draft and ending again is accepted.
	red := find(t, v.Hand, "red-7")
	l := rummy.Layout{Melds: [][]rummy.TileID{append(v.Table[0], red...)}, Hand: without(v.Hand, red)}
	if err := m.Move(rocky.ID, v.Turn, l); err != nil {
		t.Fatal(err)
	}
	ev, _ = m.EndTurn(rocky.ID, v.Turn, t0)
	if te := ended(t, ev); te.Result != rummy.ResultOK {
		t.Fatalf("TurnEnded = %+v, want ok", te)
	}
}

func TestEndTurnWithoutPlayingIsRefused(t *testing.T) {
	m := newMatch(t, rummy.Free, nil, nil)
	ev, _ := m.EndTurn(rocky.ID, 1, t0)
	if got := refused(t, ev); !slices.Equal(got, []string{"nothing_placed"}) {
		t.Fatalf("Refused = %v, want nothing_placed", got)
	}
	if m.ViewFor(bandit.ID).Current != rocky.ID {
		t.Fatal("turn passed; it should stay with rocky")
	}
}

func TestClassicOpening(t *testing.T) {
	hand := []string{"blue-3", "blue-4", "blue-5", "red-10", "blue-10", "green-10"}

	m := newMatch(t, rummy.Classic, hand, nil)
	if got := refused(t, play(t, m, rocky.ID, []string{"blue-3", "blue-4", "blue-5"})); !slices.Equal(got, []string{"opening_too_low"}) {
		t.Fatalf("12-point opening: %v, want opening_too_low (RN13)", got)
	}

	m = newMatch(t, rummy.Classic, hand, nil)
	if te := ended(t, play(t, m, rocky.ID, []string{"red-10", "blue-10", "green-10"})); te.Result != rummy.ResultOK {
		t.Fatalf("30-point opening: %+v, want ok", te)
	}

	m = newMatch(t, rummy.Free, hand, nil)
	if te := ended(t, play(t, m, rocky.ID, []string{"blue-3", "blue-4", "blue-5"})); te.Result != rummy.ResultOK {
		t.Fatalf("free mode: %+v, want ok", te)
	}
}

func TestClassicOpeningCannotUseTheTable(t *testing.T) {
	m := newMatch(t, rummy.Classic,
		[]string{"red-10", "blue-10", "green-10"},
		[]string{"brown-10", "red-11", "red-12", "red-13"})
	play(t, m, rocky.ID, []string{"red-10", "blue-10", "green-10"})

	// bandit adds a 10 to rocky's group and lays 11-12-13: 36 points, but
	// one meld takes a table tile, which the opening forbids.
	v := m.ViewFor(bandit.ID)
	ten := find(t, v.Hand, "brown-10")
	run := find(t, v.Hand, "red-11", "red-12", "red-13")
	l := rummy.Layout{
		Melds: [][]rummy.TileID{append(slices.Clone(v.Table[0]), ten...), run},
		Hand:  without(without(v.Hand, ten), run),
	}
	if err := m.Move(bandit.ID, v.Turn, l); err != nil {
		t.Fatal(err)
	}
	ev, _ := m.EndTurn(bandit.ID, v.Turn, t0)
	if got := refused(t, ev); !slices.Equal(got, []string{"opening_used_table"}) {
		t.Fatalf("Refused = %v, want opening_used_table", got)
	}
}

func TestRearrangeTheTable(t *testing.T) {
	m := newMatch(t, rummy.Free,
		[]string{"red-4", "red-5", "red-6", "red-7"},
		[]string{"blue-7", "green-7"})
	play(t, m, rocky.ID, []string{"red-4", "red-5", "red-6", "red-7"})

	v := m.ViewFor(bandit.ID)
	seven := v.Table[0][3]
	mine := find(t, v.Hand, "blue-7", "green-7")
	l := rummy.Layout{
		Melds: [][]rummy.TileID{v.Table[0][:3], append([]rummy.TileID{seven}, mine...)},
		Hand:  without(v.Hand, mine),
	}
	if err := m.Move(bandit.ID, v.Turn, l); err != nil {
		t.Fatal(err)
	}
	ev, _ := m.EndTurn(bandit.ID, v.Turn, t0)
	if te := ended(t, ev); te.Result != rummy.ResultOK {
		t.Fatalf("TurnEnded = %+v, want ok (RN11)", te)
	}
	if got := m.ViewFor(rocky.ID).Table; len(got) != 2 {
		t.Fatalf("table = %v, want the run and the group", got)
	}
}

func TestUndo(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{"blue-7", "green-7"}, nil)
	before := m.ViewFor(rocky.ID)
	ids := find(t, before.Hand, "blue-7", "green-7")
	if err := m.Move(rocky.ID, 1, rummy.Layout{Melds: [][]rummy.TileID{ids}, Hand: without(before.Hand, ids)}); err != nil {
		t.Fatal(err)
	}
	if err := m.Undo(rocky.ID, 1); err != nil {
		t.Fatal(err)
	}
	after := m.ViewFor(rocky.ID)
	if len(after.Table) != 0 || !slices.Equal(after.Hand, before.Hand) {
		t.Fatalf("after Undo: table %v, hand %v; want the start of the turn (RN12)", after.Table, after.Hand)
	}
}

func TestTimeout(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{"blue-7", "green-7"}, nil)
	if ev := m.Advance(t0.Add(39 * time.Second)); len(ev) != 0 {
		t.Fatalf("events before 40 s: %v", ev)
	}
	ev := m.Advance(t0.Add(40 * time.Second))
	if te := ended(t, ev); te.Result != rummy.ResultTimeout || !te.Drew {
		t.Fatalf("TurnEnded = %+v, want timeout with a draw (RN15)", te)
	}
	if v := m.ViewFor(rocky.ID); v.Current != bandit.ID || v.Counts[rocky.ID] != 15 {
		t.Fatalf("after timeout: turn of %s, rocky %d tiles", v.Current, v.Counts[rocky.ID])
	}
}

func TestWinning(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{
		"red-1", "red-2", "red-3", "red-4", "red-5", "red-6", "red-7", "red-8", "red-9", "red-10", "red-11",
		"green-1", "green-2", "green-3",
	}, nil)
	play(t, m, rocky.ID, []string{"red-1", "red-2", "red-3", "red-4", "red-5", "red-6", "red-7", "red-8", "red-9", "red-10", "red-11"})
	if _, err := m.Draw(bandit.ID, 2, t0); err != nil {
		t.Fatal(err)
	}
	ev := play(t, m, rocky.ID, []string{"green-1", "green-2", "green-3"})
	over, ok := m.Over()
	if !ok || over.Winner != rocky.ID {
		t.Fatalf("Over = %+v, %v; want rocky wins (RN17)", over, ok)
	}
	if _, isOver := ev[len(ev)-1].(rummy.MatchOver); !isOver {
		t.Fatalf("last event = %T, want MatchOver", ev[len(ev)-1])
	}
	if over.Scores[0].ID != rocky.ID || over.Scores[0].Points != 0 || over.Scores[1].Points == 0 {
		t.Fatalf("scores = %+v", over.Scores)
	}
}

func TestEmptyPoolEndsAfterARoundOfPasses(t *testing.T) {
	m := newMatch(t, rummy.Free, nil, nil)
	for turn := 1; m.ViewFor(rocky.ID).Pool > 0; turn++ {
		v := m.ViewFor(rocky.ID)
		if _, err := m.Draw(v.Current, v.Turn, t0); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if _, ok := m.Over(); ok {
			t.Fatal("match ended before everyone passed")
		}
		v := m.ViewFor(rocky.ID)
		ev, _ := m.Draw(v.Current, v.Turn, t0)
		if ended(t, ev).Drew {
			t.Fatal("drew from an empty pool")
		}
	}
	over, ok := m.Over()
	if !ok || len(over.Scores) != 2 {
		t.Fatalf("Over = %+v, %v; want the lowest hand to win (RN18)", over, ok)
	}
}

func TestHandPointsJokerIs30(t *testing.T) {
	if got := rummy.HandPoints([]rummy.TileID{104, 0, 12}); got != 30+1+13 {
		t.Fatalf("HandPoints = %d, want 44", got)
	}
}

func TestLeaveReturnsTilesAndPassesTheTurn(t *testing.T) {
	m := newMatch(t, rummy.Free, nil, nil, nil)
	ev := m.Leave(rocky.ID, t0)
	if _, ok := ev[0].(rummy.PlayerRemoved); !ok {
		t.Fatalf("first event = %T, want PlayerRemoved", ev[0])
	}
	v := m.ViewFor(bandit.ID)
	if v.Pool != 106-3*14+14 || v.Current != bandit.ID || m.Has(rocky.ID) {
		t.Fatalf("pool %d, turn of %s; want 78 and bandit (RN20)", v.Pool, v.Current)
	}

	m.Leave(mel.ID, t0)
	if over, ok := m.Over(); !ok || over.Winner != bandit.ID {
		t.Fatalf("Over = %+v, %v; want bandit wins alone (RN22)", over, ok)
	}
}

func TestLeaveBeforeTheCurrentPlayerKeepsTheTurn(t *testing.T) {
	m := newMatch(t, rummy.Free, nil, nil, nil)
	if _, err := m.Draw(rocky.ID, 1, t0); err != nil {
		t.Fatal(err)
	}
	m.Leave(rocky.ID, t0)
	if v := m.ViewFor(bandit.ID); v.Current != bandit.ID || v.Turn != 2 {
		t.Fatalf("turn %d of %s, want 2 of bandit", v.Turn, v.Current)
	}
}

func TestAbsentSeatIsKeptThenFreed(t *testing.T) {
	m := newMatch(t, rummy.Free, nil, nil, nil)
	hand := m.ViewFor(bandit.ID).Hand
	m.SetAbsent(bandit.ID, t0)

	old, ok := m.Reclaim("bandit", "p9")
	if !ok || old != bandit.ID {
		t.Fatalf("Reclaim = %s, %v", old, ok)
	}
	if got := m.ViewFor("p9").Hand; !slices.Equal(got, hand) {
		t.Fatal("came back with a different hand (RN21)")
	}

	m.SetAbsent(mel.ID, t0)
	if ev := m.Advance(t0.Add(29 * time.Second)); len(ev) != 0 {
		t.Fatalf("events before 30 s: %v", ev)
	}
	m.Advance(t0.Add(30 * time.Second))
	if m.Has(mel.ID) {
		t.Fatal("mel still seated after 30 s away")
	}
	if _, ok := m.Reclaim("mel", "p8"); ok {
		t.Fatal("reclaimed a freed seat")
	}
}

func TestMeldPositions(t *testing.T) {
	m := newMatch(t, rummy.Free, []string{"blue-3", "blue-4", "blue-5"}, nil)
	v := m.ViewFor(rocky.ID)
	ids := find(t, v.Hand, "blue-3", "blue-4", "blue-5")
	l := rummy.Layout{Melds: [][]rummy.TileID{ids}, Pos: []rummy.Pos{{X: 20, Y: 9}}, Hand: without(v.Hand, ids)}
	if err := m.Move(rocky.ID, v.Turn, l); err != nil {
		t.Fatal(err)
	}
	if _, _, pos, _ := m.Draft(); !slices.Equal(pos, []rummy.Pos{{X: 20, Y: 9}}) {
		t.Fatalf("draft positions = %v, want the ones sent (RN32)", pos)
	}
	l.Pos = []rummy.Pos{{X: rummy.TableCols, Y: 0}}
	if err := m.Move(rocky.ID, v.Turn, l); !errors.Is(err, rummy.ErrBadLayout) {
		t.Fatalf("Move off the table = %v, want ErrBadLayout", err)
	}
	m.EndTurn(rocky.ID, v.Turn, t0)
	if got := m.ViewFor(bandit.ID).Pos; !slices.Equal(got, []rummy.Pos{{X: 20, Y: 9}}) {
		t.Fatalf("table positions = %v, want them kept after the turn", got)
	}
}
