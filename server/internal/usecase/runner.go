package usecase

import (
	"context"
	"errors"
	"time"

	"raccoons-world/server/internal/domain/cups"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/domain/rummy"
)

// TickInterval is the simulation step: 20 ticks per second. Every tick
// advances movement by exactly this amount, regardless of scheduling jitter,
// so the simulation stays deterministic.
const TickInterval = 50 * time.Millisecond

// ErrRoomStopped is returned by Exec once the room loop has ended.
var ErrRoomStopped = errors.New("room stopped")

type command struct {
	fn   func()
	done chan struct{}
}

// RoomRunner owns a room, its three cups booth and its pieces table, and serializes all access
// to them through a single goroutine (Run). Commands and ticks never run
// concurrently, so neither needs locks.
type RoomRunner struct {
	room        *room.Room
	game        *cups.Game
	table       *rummy.Table
	broadcaster Broadcaster
	clock       Clock
	cmds        chan command
	stopped     chan struct{}
	tick        uint64                 // only touched by the loop goroutine
	cupsBusy    map[room.PlayerID]bool // players the booth holds still (RC21)
	rummyBusy   map[room.PlayerID]bool // players the table holds still (RN27)
}

// NewRoomRunner wraps r, its booth g and its pieces table t. Call Run to start processing. c is
// used on each tick to expire silent players (RN5) and to advance the match.
func NewRoomRunner(r *room.Room, g *cups.Game, t *rummy.Table, b Broadcaster, c Clock) *RoomRunner {
	return &RoomRunner{
		room:        r,
		game:        g,
		table:       t,
		broadcaster: b,
		clock:       c,
		cmds:        make(chan command),
		stopped:     make(chan struct{}),
		cupsBusy:    map[room.PlayerID]bool{},
		rummyBusy:   map[room.PlayerID]bool{},
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
			c.fn()
			close(c.done)
		case <-ticks:
			rr.step()
		}
	}
}

// Exec runs fn on the room goroutine and waits for it to finish. fn must
// not call Exec (that would deadlock) and must not keep the *room.Room.
func (rr *RoomRunner) Exec(fn func(*room.Room)) error {
	return rr.do(func() { fn(rr.room) })
}

// do runs fn on the room goroutine; use cases in this package reach the
// room and the booth through it.
func (rr *RoomRunner) do(fn func()) error {
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
	now := rr.clock.Now()
	for _, p := range rr.room.Expire(now) {
		rr.leaveBooth(p.ID, now)
		rr.dropFromTable(p.ID, now)
		rr.broadcaster.ToRoom(rr.room.ID(), PlayerLeft{ID: p.ID})
	}
	rr.publishCups(rr.game.Advance(now))
	rr.publishRummy(rr.table.Advance(now), now)

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

// leaveBooth takes a player who left the room out of the match (RC15).
func (rr *RoomRunner) leaveBooth(id room.PlayerID, now time.Time) {
	_, events := rr.game.Leave(id, now)
	rr.publishCups(events)
}

// publishCups sends booth events to the room and keeps the busy flags in
// step with who is at the booth. Picks never reveal the cup (RC10).
func (rr *RoomRunner) publishCups(events []cups.Event) {
	id := rr.room.ID()
	for _, e := range events {
		switch e := e.(type) {
		case cups.StatusChanged:
			rr.syncBusy(e)
			rr.broadcaster.ToRoom(id, CupsStatus{Status: e})
		case cups.RoundStarted:
			rr.broadcaster.ToRoom(id, CupsRound{Round: e})
		case cups.Picked:
			rr.broadcaster.ToRoom(id, CupsPicked{ID: e.ID})
		case cups.Revealed:
			rr.broadcaster.ToRoom(id, CupsRevealed{Reveal: e})
		case cups.Finished:
			rr.broadcaster.ToRoom(id, CupsResults{Results: e})
		}
	}
}

func (rr *RoomRunner) syncBusy(s cups.StatusChanged) {
	now := map[room.PlayerID]bool{}
	for _, p := range s.Players {
		now[p] = true
	}
	for _, p := range s.Queued {
		now[p] = true
	}
	before := rr.cupsBusy
	rr.cupsBusy = now
	rr.applyBusy(before, now)
}

// applyBusy updates the room for everyone in either of one minigame's busy
// sets: a player is held still while any minigame holds them.
func (rr *RoomRunner) applyBusy(before, now map[room.PlayerID]bool) {
	for _, set := range []map[room.PlayerID]bool{before, now} {
		for p := range set {
			if before[p] != now[p] {
				// They may have left the room already.
				_ = rr.room.SetBusy(p, rr.cupsBusy[p] || rr.rummyBusy[p])
			}
		}
	}
}

// cupsSnapshot is the booth as a newcomer to the room must see it.
func (rr *RoomRunner) cupsSnapshot(now time.Time) CupsSnapshot {
	s := CupsSnapshot{Status: rr.game.Status(now)}
	if r, elapsed, ok := rr.game.CurrentRound(now); ok {
		s.Round = &CupsRound{Round: r, Elapsed: elapsed}
	}
	return s
}
