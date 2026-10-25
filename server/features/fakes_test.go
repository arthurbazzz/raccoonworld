package features

import (
	"sync"
	"time"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

var (
	_ usecase.Clock       = (*fakeClock)(nil)
	_ usecase.Broadcaster = (*fakeBroadcaster)(nil)
)

// fakeClock is a manually advanced clock so time-based rules (chat rate
// limit, disconnect timeout) run instantly and deterministically.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// sentEvent is one message the broadcaster was asked to deliver.
type sentEvent struct {
	RoomID   string        // set for room-wide broadcasts
	PlayerID room.PlayerID // set for direct messages
	Event    usecase.Event
}

// fakeBroadcaster records every event so steps can assert who saw what.
type fakeBroadcaster struct {
	mu     sync.Mutex
	events []sentEvent
}

func (b *fakeBroadcaster) ToRoom(roomID string, ev usecase.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, sentEvent{RoomID: roomID, Event: ev})
}

func (b *fakeBroadcaster) ToPlayer(playerID room.PlayerID, ev usecase.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, sentEvent{PlayerID: playerID, Event: ev})
}

func (b *fakeBroadcaster) Events() []sentEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]sentEvent(nil), b.events...)
}
