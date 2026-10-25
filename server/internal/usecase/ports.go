// Package usecase orchestrates the room domain: it runs each room on its own
// goroutine and turns domain results into events for the outside world.
// It defines the ports (interfaces) adapters must implement.
package usecase

import (
	"time"

	"raccoons-world/server/internal/domain/room"
)

// Clock is the source of wall-clock time. Tests use a fake one.
type Clock interface {
	Now() time.Time
}

// Broadcaster delivers events to connected players. The WebSocket adapter
// implements it; tests use a recording fake.
type Broadcaster interface {
	// ToRoom sends ev to every player in the room.
	ToRoom(roomID string, ev Event)
	// ToPlayer sends ev to a single player.
	ToPlayer(playerID room.PlayerID, ev Event)
}

// Event is something that happened in a room and must reach players.
// The set is closed: only types in this package implement it.
type Event interface {
	isEvent()
}

// Snapshot carries the new positions of the players that moved in a tick.
type Snapshot struct {
	RoomID  string
	Tick    uint64
	Players []room.PlayerView
}

func (Snapshot) isEvent() {}
