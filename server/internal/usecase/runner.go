package usecase

import (
	"context"
	"errors"
	"time"

	"raccoons-world/server/internal/domain/room"
)

// TickInterval is the simulation step: 20 ticks per second. Every tick
// advances movement by exactly this amount, regardless of scheduling jitter,
// so the simulation stays deterministic.
const TickInterval = 50 * time.Millisecond

// ErrRoomStopped is returned by Exec once the room loop has ended.
var ErrRoomStopped = errors.New("room stopped")

type command struct {
	fn   func(*room.Room)
	done chan struct{}
}

// RoomRunner owns a room and serializes all access to it through a single
// goroutine (Run). Commands and ticks never run concurrently, so the room
// needs no locks.
type RoomRunner struct {
	room        *room.Room
	broadcaster Broadcaster
	clock       Clock
	cmds        chan command
	stopped     chan struct{}
	tick        uint64 // only touched by the loop goroutine
}

// NewRoomRunner wraps r. Call Run to start processing. c is used on each
// tick to expire silent players (RN5).
func NewRoomRunner(r *room.Room, b Broadcaster, c Clock) *RoomRunner {
	return &RoomRunner{
		room:        r,
		broadcaster: b,
		clock:       c,
		cmds:        make(chan command),
		stopped:     make(chan struct{}),
	}
}

// RoomID returns the identifier of the wrapped room.
func (rr *RoomRunner) RoomID() string { return rr.room.ID() }

// Run processes commands and ticks until ctx is done. It must be called
// exactly once. In production ticks comes from a time.Ticker with
// TickInterval; tests drive it by hand.
func (rr *RoomRunner) Run(ctx context.Context, ticks <-chan time.Time) {
	defer close(rr.stopped)
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-rr.cmds:
			c.fn(rr.room)
			close(c.done)
		case <-ticks:
			rr.step()
		}
	}
}

// Exec runs fn on the room goroutine and waits for it to finish. fn must
// not call Exec (that would deadlock) and must not keep the *room.Room.
func (rr *RoomRunner) Exec(fn func(*room.Room)) error {
	c := command{fn: fn, done: make(chan struct{})}
	select {
	case rr.cmds <- c:
	case <-rr.stopped:
		return ErrRoomStopped
	}
	<-c.done
	return nil
}

func (rr *RoomRunner) step() {
	rr.tick++
	for _, p := range rr.room.Expire(rr.clock.Now()) {
		rr.broadcaster.ToRoom(rr.room.ID(), PlayerLeft{ID: p.ID})
	}
	moved := rr.room.Step(TickInterval)
	if len(moved) == 0 {
		return
	}
	rr.broadcaster.ToRoom(rr.room.ID(), Snapshot{
		RoomID:  rr.room.ID(),
		Tick:    rr.tick,
		Players: moved,
	})
}
