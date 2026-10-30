package rummy_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/rummy"
)

var (
	pipoca = rummy.Player{ID: "p4", Nickname: "pipoca"}
	tico   = rummy.Player{ID: "p5", Nickname: "tico"}
)

func newTable() *rummy.Table {
	return rummy.NewTable(rummy.DefaultTiming(), rummy.NewRandomShuffler(1))
}

func mustCreate(t *testing.T, tb *rummy.Table, p rummy.Player, mode rummy.Mode) string {
	t.Helper()
	id, _, err := tb.Create(p, mode)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustJoin(t *testing.T, tb *rummy.Table, p rummy.Player, roomID string) {
	t.Helper()
	if _, err := tb.Join(p, roomID, t0); err != nil {
		t.Fatal(err)
	}
}

func roomInfo(t *testing.T, tb *rummy.Table, id string) rummy.RoomInfo {
	t.Helper()
	for _, r := range tb.Rooms() {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("no room %s", id)
	return rummy.RoomInfo{}
}

func TestCreateRoom(t *testing.T) {
	tb := newTable()
	id := mustCreate(t, tb, rocky, rummy.Classic)
	r := roomInfo(t, tb, id)
	if r.Mode != rummy.Classic || len(r.Players) != 1 || r.Phase != rummy.Lobby || !r.LobbyEnds.IsZero() {
		t.Fatalf("room = %+v, want a classic lobby with rocky and no countdown", r)
	}
	if _, _, err := tb.Create(rocky, rummy.Free); !errors.Is(err, rummy.ErrSeated) {
		t.Fatalf("second Create = %v, want ErrSeated", err)
	}
	if _, _, err := tb.Create(bandit, 0); !errors.Is(err, rummy.ErrNoMode) {
		t.Fatalf("Create without a mode = %v", err)
	}
}

func TestCountdownStartsWithTwoAndStopsWithOne(t *testing.T) {
	tb := newTable()
	id := mustCreate(t, tb, rocky, rummy.Free)
	mustJoin(t, tb, bandit, id)
	if got := roomInfo(t, tb, id).LobbyEnds; !got.Equal(t0.Add(15 * time.Second)) {
		t.Fatalf("LobbyEnds = %v, want 15 s from now (RN6)", got)
	}
	tb.Leave(bandit.ID, t0)
	if !roomInfo(t, tb, id).LobbyEnds.IsZero() {
		t.Fatal("countdown kept going with one player")
	}
	mustJoin(t, tb, bandit, id)
	tb.Advance(t0.Add(15 * time.Second))
	if r := roomInfo(t, tb, id); r.Phase != rummy.Playing {
		t.Fatalf("phase = %v after the countdown, want playing", r.Phase)
	}
}

func TestFourPlayersStartAtOnce(t *testing.T) {
	tb := newTable()
	id := mustCreate(t, tb, rocky, rummy.Free)
	mustJoin(t, tb, bandit, id)
	mustJoin(t, tb, mel, id)
	ev, err := tb.Join(pipoca, id, t0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(ev, func(e rummy.Event) bool { _, ok := e.(rummy.MatchStarted); return ok }) {
		t.Fatalf("events = %v, want MatchStarted (RN6)", ev)
	}
	if _, err := tb.Join(tico, id, t0); !errors.Is(err, rummy.ErrPlaying) {
		t.Fatalf("Join a playing room = %v, want ErrPlaying (RN7)", err)
	}
}

func TestTwoRoomsPlayAtTheSameTime(t *testing.T) {
	tb := newTable()
	a := mustCreate(t, tb, rocky, rummy.Free)
	mustJoin(t, tb, bandit, a)
	b := mustCreate(t, tb, mel, rummy.Classic)
	mustJoin(t, tb, pipoca, b)
	tb.Advance(t0.Add(15 * time.Second))
	if tb.Matches() != 2 {
		t.Fatalf("Matches = %d, want 2", tb.Matches())
	}
	ma, _ := tb.Match(a)
	if ma.Has(mel.ID) {
		t.Fatal("rooms share a match")
	}
}

func TestLastToLeaveRemovesTheRoom(t *testing.T) {
	tb := newTable()
	id := mustCreate(t, tb, rocky, rummy.Free)
	tb.Leave(rocky.ID, t0)
	if len(tb.Rooms()) != 0 {
		t.Fatalf("rooms = %v, want none (RN5)", tb.Rooms())
	}
	if _, err := tb.Join(bandit, id, t0); !errors.Is(err, rummy.ErrNoRoom) {
		t.Fatalf("Join a removed room = %v", err)
	}
}

func TestPlayAgainKeepsTheRoom(t *testing.T) {
	tb := newTable()
	id := mustCreate(t, tb, rocky, rummy.Free)
	mustJoin(t, tb, bandit, id)
	mustJoin(t, tb, mel, id)
	tb.Advance(t0.Add(15 * time.Second))
	tb.Leave(mel.ID, t0.Add(16*time.Second))
	tb.Leave(bandit.ID, t0.Add(16*time.Second)) // rocky wins alone (RN22)
	if r := roomInfo(t, tb, id); r.Phase != rummy.Results {
		t.Fatalf("phase = %v, want results", r.Phase)
	}
	tb.PlayAgain(rocky.ID)
	ev := tb.Advance(t0.Add(30 * time.Second))
	if r := roomInfo(t, tb, id); r.Phase != rummy.Lobby || len(r.Players) != 1 {
		t.Fatalf("room = %+v, want rocky back in the lobby (RN7)", r)
	}
	if slices.ContainsFunc(ev, func(e rummy.Event) bool { _, ok := e.(rummy.Released); return ok }) {
		t.Fatal("rocky was released though he asked to play again")
	}
}

func TestDroppedInMatchKeepsTheSeat(t *testing.T) {
	tb := newTable()
	id := mustCreate(t, tb, rocky, rummy.Free)
	mustJoin(t, tb, bandit, id)
	tb.Advance(t0.Add(15 * time.Second))
	tb.SetAbsent(rocky.ID, t0.Add(16*time.Second))
	roomID, old, ok := tb.Reclaim("rocky", "p9")
	if !ok || roomID != id || old != rocky.ID {
		t.Fatalf("Reclaim = %s, %s, %v", roomID, old, ok)
	}
	if got, _ := tb.RoomOf("p9"); got != id {
		t.Fatal("came back outside the room")
	}
}

func TestDroppedInLobbyLeaves(t *testing.T) {
	tb := newTable()
	mustCreate(t, tb, rocky, rummy.Free)
	tb.SetAbsent(rocky.ID, t0)
	if len(tb.Rooms()) != 0 {
		t.Fatal("a dropped player kept a lobby")
	}
}
