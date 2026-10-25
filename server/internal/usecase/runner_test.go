package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
	"raccoons-world/server/maps"
)

// epoch is the time every test player joins; fixedClock never moves past it,
// so nobody times out in these tests.
var epoch = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return epoch }

type recordingBroadcaster struct {
	mu     sync.Mutex
	toRoom []usecase.Event
}

func (b *recordingBroadcaster) ToRoom(_ string, ev usecase.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.toRoom = append(b.toRoom, ev)
}

func (b *recordingBroadcaster) ToPlayer(room.PlayerID, usecase.Event) {}

func (b *recordingBroadcaster) roomEvents() []usecase.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]usecase.Event(nil), b.toRoom...)
}

// startRunner runs a Praça room loop driven by a manual tick channel.
func startRunner(t *testing.T) (*usecase.RoomRunner, chan time.Time, *recordingBroadcaster) {
	t.Helper()
	tiles, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		t.Fatal(err)
	}
	b := &recordingBroadcaster{}
	runner := usecase.NewRoomRunner(room.New("praca", tiles), b, fixedClock{})
	ticks := make(chan time.Time)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		runner.Run(ctx, ticks)
	}()
	t.Cleanup(func() { cancel(); <-done })
	return runner, ticks, b
}

// tick sends one tick and waits until the loop has processed it.
func tick(t *testing.T, r *usecase.RoomRunner, ticks chan<- time.Time) {
	t.Helper()
	ticks <- time.Time{}
	if err := r.Exec(func(*room.Room) {}); err != nil { // barrier
		t.Fatal(err)
	}
}

func TestExecRunsOnRoomLoop(t *testing.T) {
	r, _, _ := startRunner(t)
	nick, _ := room.NewNickname("rocky")

	var joinErr error
	if err := r.Exec(func(rm *room.Room) { _, joinErr = rm.Join("p1", nick, epoch) }); err != nil {
		t.Fatal(err)
	}
	if joinErr != nil {
		t.Fatal(joinErr)
	}

	var count int
	_ = r.Exec(func(rm *room.Room) { count = len(rm.Players()) })
	if count != 1 {
		t.Fatalf("players = %d, want 1", count)
	}
}

func TestTickBroadcastsSnapshotOfMovedPlayers(t *testing.T) {
	r, ticks, b := startRunner(t)
	nick, _ := room.NewNickname("rocky")
	_ = r.Exec(func(rm *room.Room) {
		_, _ = rm.Join("p1", nick, epoch)
		_ = rm.SetInput("p1", room.Input{Right: true})
	})

	tick(t, r, ticks)
	tick(t, r, ticks)

	events := b.roomEvents()
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2 snapshots", len(events))
	}
	snap, ok := events[1].(usecase.Snapshot)
	if !ok {
		t.Fatalf("event = %T, want Snapshot", events[1])
	}
	if snap.Tick != 2 || len(snap.Players) != 1 || snap.Players[0].ID != "p1" {
		t.Fatalf("snapshot = %+v", snap)
	}
	wantX := 480 + 2*room.Speed*usecase.TickInterval.Seconds()
	if got := snap.Players[0].Position.X; got != wantX {
		t.Fatalf("X = %v, want %v", got, wantX)
	}
}

func TestIdleTickBroadcastsNothing(t *testing.T) {
	r, ticks, b := startRunner(t)
	nick, _ := room.NewNickname("rocky")
	_ = r.Exec(func(rm *room.Room) { _, _ = rm.Join("p1", nick, epoch) })

	tick(t, r, ticks)

	if n := len(b.roomEvents()); n != 0 {
		t.Fatalf("events = %d, want 0 when nobody moves", n)
	}
}

func TestExecAfterStopFails(t *testing.T) {
	tiles, _ := room.ParseTileMap(maps.Praca)
	r := usecase.NewRoomRunner(room.New("praca", tiles), &recordingBroadcaster{}, fixedClock{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); r.Run(ctx, nil) }()
	cancel()
	<-done

	if err := r.Exec(func(*room.Room) {}); !errors.Is(err, usecase.ErrRoomStopped) {
		t.Fatalf("Exec() error = %v, want ErrRoomStopped", err)
	}
}

func TestLeaveRoomBroadcastsOnce(t *testing.T) {
	r, _, b := startRunner(t)
	nick, _ := room.NewNickname("rocky")
	_ = r.Exec(func(rm *room.Room) { _, _ = rm.Join("p1", nick, epoch) })
	leave := usecase.NewLeaveRoom(r, b)

	for range 2 { // a clean close after a timeout must not announce twice
		if err := leave.Execute("p1"); err != nil {
			t.Fatal(err)
		}
	}

	events := b.roomEvents()
	if len(events) != 1 || events[0] != (usecase.PlayerLeft{ID: "p1"}) {
		t.Fatalf("events = %+v, want one PlayerLeft{p1}", events)
	}
}

func TestKeepAliveUnknownPlayer(t *testing.T) {
	r, _, _ := startRunner(t)
	err := usecase.NewKeepAlive(r, fixedClock{}).Execute("ghost")
	if !errors.Is(err, room.ErrPlayerNotFound) {
		t.Fatalf("Execute() error = %v, want room.ErrPlayerNotFound", err)
	}
}

func TestTickIntervalIs20Hz(t *testing.T) {
	if usecase.TickInterval != 50*time.Millisecond {
		t.Fatalf("TickInterval = %v, want 50ms", usecase.TickInterval)
	}
}
