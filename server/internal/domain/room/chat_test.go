package room_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/room"
)

var t0 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestSayReturnsMessage(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")

	msg, err := r.Say("p1", "olá!", t0)
	if err != nil {
		t.Fatalf("Say() error: %v", err)
	}
	if msg.From != "p1" || msg.Text != "olá!" || !msg.SentAt.Equal(t0) {
		t.Fatalf("Say() = %+v", msg)
	}
}

func TestSayLengthLimitCountsCharactersNotBytes(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")

	// 120 accented characters are 240 bytes but still within RN4.
	if _, err := r.Say("p1", strings.Repeat("é", room.MaxMessageLength), t0); err != nil {
		t.Fatalf("Say(120 chars) error: %v", err)
	}
	_, err := r.Say("p1", strings.Repeat("a", room.MaxMessageLength+1), t0.Add(time.Minute))
	if !errors.Is(err, room.ErrMessageTooLong) {
		t.Fatalf("Say(121 chars) error = %v, want ErrMessageTooLong", err)
	}
	if room.MaxMessageLength != 120 {
		t.Fatalf("MaxMessageLength = %d, want 120 (RN4)", room.MaxMessageLength)
	}
}

func TestSayRejectsBlankMessages(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	for i, text := range []string{"", "   ", "\n\t"} {
		_, err := r.Say("p1", text, t0.Add(time.Duration(i)*time.Minute))
		if !errors.Is(err, room.ErrMessageEmpty) {
			t.Fatalf("Say(%q) error = %v, want ErrMessageEmpty", text, err)
		}
	}
}

func TestSayRateLimit(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")

	if _, err := r.Say("p1", "first", t0); err != nil {
		t.Fatal(err)
	}
	_, err := r.Say("p1", "too soon", t0.Add(999*time.Millisecond))
	if !errors.Is(err, room.ErrRateLimited) {
		t.Fatalf("Say() after 999ms error = %v, want ErrRateLimited", err)
	}
	// Rejected messages do not reset the window: 1s after the first is fine.
	if _, err := r.Say("p1", "on time", t0.Add(time.Second)); err != nil {
		t.Fatalf("Say() after 1s error: %v", err)
	}
}

func TestSayRateLimitIsPerPlayer(t *testing.T) {
	r := newTestRoom(t)
	mustJoin(t, r, "p1", "rocky")
	mustJoin(t, r, "p2", "bandit")

	if _, err := r.Say("p1", "oi", t0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Say("p2", "oi", t0); err != nil {
		t.Fatalf("p2 limited by p1's message: %v", err)
	}
}

func TestSayUnknownPlayer(t *testing.T) {
	r := newTestRoom(t)
	if _, err := r.Say("ghost", "boo", t0); !errors.Is(err, room.ErrPlayerNotFound) {
		t.Fatalf("Say() error = %v, want ErrPlayerNotFound", err)
	}
}
