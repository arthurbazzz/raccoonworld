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

// Joined is sent to a player who entered a room: who they are and who is
// there, themselves included.
type Joined struct {
	RoomID  string
	MapName string
	You     room.PlayerView
	Players []room.PlayerView
}

// JoinRejected is sent to a visitor whose join attempt failed.
type JoinRejected struct {
	Reason RejectReason
}

// PlayerJoined is broadcast to the room when someone enters. The newcomer
// receives it too; clients ignore the event for their own ID.
type PlayerJoined struct {
	Player room.PlayerView
}

// PlayerLeft is broadcast to the room when someone leaves or times out.
type PlayerLeft struct {
	ID room.PlayerID
}

// ChatMessage is broadcast to the room when a chat line is accepted.
type ChatMessage struct {
	From room.PlayerID
	Text string
}

func (Snapshot) isEvent()     {}
func (ChatMessage) isEvent()  {}
func (PlayerLeft) isEvent()   {}
func (Joined) isEvent()       {}
func (JoinRejected) isEvent() {}
func (PlayerJoined) isEvent() {}
