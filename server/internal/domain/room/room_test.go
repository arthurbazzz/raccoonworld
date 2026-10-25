package room_test

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/room"
)

const tick = 50 * time.Millisecond

func newTestRoom(t *testing.T) *room.Room {
	t.Helper()
	return room.New("praca", mustParse(t, testMap))
}

func nick(t *testing.T, s string) room.Nickname {
	t.Helper()
	n, err := room.NewNickname(s)
	if err != nil {
		t.Fatalf("NewNickname(%q): %v", s, err)
	}
	return n
}

func mustJoin(t *testing.T, r *room.Room, id, name string) room.PlayerView {
	t.Helper()
	p, err := r.Join(room.PlayerID(id), nick(t, name), t0)
	if err != nil {
		t.Fatalf("Join(%s): %v", name, err)
	}
	return p
}

func position(t *testing.T, r *room.Room, id string) room.Vec {
	t.Helper()
	p, ok := r.Player(room.PlayerID(id))
	if !ok {
		t.Fatalf("player %s not in room", id)
	}
	return p.Position
}

func TestJoinPlacesPlayerAtSpawn(t *testing.T) {
	r := newTestRoom(t)
	p := mustJoin(t, r, "p1", "rocky")

	if p.ID != "p1" || p.Nickname != "rocky" || p.Position != (room.Vec{X: 64, Y: 80}) {
		t.Fatalf("Join() = %+v", p)
	}
	if got := r.Players(); len(got) != 1 || got[0] != p {
		t.Fatalf("Players() = %+v", got)
	}
}

func TestJoinRejectsNicknameInUseIgnoringCase(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")

	for _, name := range []string{"rocky", "Rocky", "ROCKY"} {
		_, err := r.Join("p2", nick(t, name), t0)
		if !errors.Is(err, room.ErrNicknameTaken) {
			t.Fatalf("Join(%q) error = %v, want ErrNicknameTaken", name, err)
		}
	}
}

func TestJoinRejectsWhenFull(t *testing.T) {
	r := newTestRoom(t)
	for i := range room.MaxPlayers {
		mustJoin(t, r, fmt.Sprintf("p%d", i), fmt.Sprintf("player%d", i))
	}

	_, err := r.Join("extra", nick(t, "bandit"), t0)
	if !errors.Is(err, room.ErrRoomFull) {
		t.Fatalf("Join() error = %v, want ErrRoomFull", err)
	}
	if room.MaxPlayers != 10 {
		t.Fatalf("MaxPlayers = %d, want 10 (RN2)", room.MaxPlayers)
	}
}

func TestJoinRejectsDuplicateID(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	if _, err := r.Join("p1", nick(t, "bandit"), t0); !errors.Is(err, room.ErrPlayerExists) {
		t.Fatalf("Join() error = %v, want ErrPlayerExists", err)
	}
}

func TestLeaveFreesSlotAndNickname(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")

	if !r.Leave("p1") {
		t.Fatal("Leave(p1) = false, want true")
	}
	if r.Leave("p1") {
		t.Fatal("second Leave(p1) = true, want false")
	}
	mustJoin(t, r, "p2", "Rocky") // nickname is free again
}

func TestPlayersKeepsJoinOrder(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "a", "alpha")
	mustJoin(t, r, "b", "bravo")
	mustJoin(t, r, "c", "charlie")
	r.Leave("b")

	got := r.Players()
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "c" {
		t.Fatalf("Players() = %+v, want [a c]", got)
	}
}

func TestSetInputUnknownPlayer(t *testing.T) {
	r := newTestRoom(t)
	if err := r.SetInput("ghost", room.Input{Right: true}); !errors.Is(err, room.ErrPlayerNotFound) {
		t.Fatalf("SetInput() error = %v, want ErrPlayerNotFound", err)
	}
}

func TestStepMovesAtFixedSpeed(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	if err := r.SetInput("p1", room.Input{Right: true}); err != nil {
		t.Fatal(err)
	}

	r.Step(tick)

	want := room.Vec{X: 64 + room.Speed*tick.Seconds(), Y: 80}
	if got := position(t, r, "p1"); got != want {
		t.Fatalf("position = %v, want %v", got, want)
	}
}

func TestStepDiagonalSameSpeed(t *testing.T) {
	r := newTestRoom(t)
	start := mustJoin(t, r, "p1", "rocky").Position
	if err := r.SetInput("p1", room.Input{Up: true, Right: true}); err != nil {
		t.Fatal(err)
	}

	r.Step(tick)

	got := position(t, r, "p1")
	dist := math.Hypot(got.X-start.X, got.Y-start.Y)
	if math.Abs(dist-room.Speed*tick.Seconds()) > 1e-9 {
		t.Fatalf("diagonal distance = %v, want %v", dist, room.Speed*tick.Seconds())
	}
	if got.X <= start.X || got.Y >= start.Y {
		t.Fatalf("position = %v, want up-right of %v", got, start)
	}
}

func TestStepStopsWhenKeysReleased(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	_ = r.SetInput("p1", room.Input{Right: true})
	r.Step(tick)
	_ = r.SetInput("p1", room.Input{})
	before := position(t, r, "p1")

	r.Step(tick)

	if got := position(t, r, "p1"); got != before {
		t.Fatalf("position = %v, want unchanged %v", got, before)
	}
}

func TestStepCollidesWithWalls(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	_ = r.SetInput("p1", room.Input{Left: true})

	for range 20 { // 1s of holding A: far more than the distance to the wall
		r.Step(tick)
	}

	// Left border ends at x=32; the hitbox is centered on the feet.
	want := 32 + room.HitboxWidth/2
	if got := position(t, r, "p1"); got.X != want {
		t.Fatalf("X = %v, want %v (flush with the wall)", got.X, want)
	}
}

func TestStepReportsOnlyPlayersThatMoved(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	mustJoin(t, r, "p2", "bandit")
	_ = r.SetInput("p1", room.Input{Right: true})

	moved := r.Step(tick)
	if len(moved) != 1 || moved[0].ID != "p1" {
		t.Fatalf("Step() = %+v, want only p1", moved)
	}

	_ = r.SetInput("p1", room.Input{})
	if moved := r.Step(tick); len(moved) != 0 {
		t.Fatalf("Step() = %+v, want nothing when idle", moved)
	}
}
