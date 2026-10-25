package room_test

import (
	"errors"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/room"
)

func TestExpireRemovesPlayersSilentForTheTimeout(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky") // joined at t0

	if gone := r.Expire(t0.Add(room.PresenceTimeout - time.Millisecond)); len(gone) != 0 {
		t.Fatalf("Expire() before timeout = %+v, want none", gone)
	}
	gone := r.Expire(t0.Add(room.PresenceTimeout))
	if len(gone) != 1 || gone[0].ID != "p1" {
		t.Fatalf("Expire() at timeout = %+v, want [p1]", gone)
	}
	if _, ok := r.Player("p1"); ok {
		t.Fatal("expired player still in the room")
	}
	if room.PresenceTimeout != 10*time.Second {
		t.Fatalf("PresenceTimeout = %v, want 10s (RN5)", room.PresenceTimeout)
	}
}

func TestTouchKeepsPlayerAlive(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	mustJoin(t, r, "p2", "bandit")

	if err := r.Touch("p1", t0.Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}

	gone := r.Expire(t0.Add(room.PresenceTimeout))
	if len(gone) != 1 || gone[0].ID != "p2" {
		t.Fatalf("Expire() = %+v, want only the silent p2", gone)
	}
	if gone := r.Expire(t0.Add(5*time.Second + room.PresenceTimeout)); len(gone) != 1 || gone[0].ID != "p1" {
		t.Fatalf("Expire() later = %+v, want p1", gone)
	}
}

func TestTouchUnknownPlayer(t *testing.T) {
	r := newTestRoom(t)
	if err := r.Touch("ghost", t0); !errors.Is(err, room.ErrPlayerNotFound) {
		t.Fatalf("Touch() error = %v, want ErrPlayerNotFound", err)
	}
}
