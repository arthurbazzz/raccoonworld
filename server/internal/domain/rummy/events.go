package rummy

import (
	"time"

	"raccoons-world/server/internal/domain/room"
)

// Event is something that happened at the table, for the use cases to
// broadcast.
type Event interface{ rummyEvent() }

// TurnStarted says it is now Player's turn, until Ends (RN15).
type TurnStarted struct {
	Turn   int
	Player room.PlayerID
	Ends   time.Time
}

// TurnEnded says how the turn ended. Drew is false when the pool was empty
// (RN16).
type TurnEnded struct {
	Turn   int
	By     room.PlayerID
	Result Result
	Drew   bool
}

// Refused says the current player's play was refused and why (RN12a, RN14).
// The turn goes on: only that player needs to know.
type Refused struct {
	Turn   int
	Player room.PlayerID
	Issues []Issue
}

// PlayerRemoved says a player left the match or did not come back (RN20, RN21).
type PlayerRemoved struct {
	Player Player
}

// Score is one line of the final scoreboard (RN19).
type Score struct {
	Player
	Points int // left in the hand
	Place  int // 1-based; ties share the place
}

// MatchOver says the match ended. Winner is empty on a tie (RN18).
type MatchOver struct {
	Winner room.PlayerID
	Scores []Score
}

func (TurnStarted) rummyEvent()   {}
func (TurnEnded) rummyEvent()     {}
func (Refused) rummyEvent()       {}
func (PlayerRemoved) rummyEvent() {}
func (MatchOver) rummyEvent()     {}

// RoomsChanged says the room list changed (RN5-RN7).
type RoomsChanged struct{}

// MatchStarted says a room started a match with these players, in turn order.
type MatchStarted struct {
	Room    string
	Players []Player
}

// InMatch is a match event of one room. Event is nil when only the match
// state changed (someone dropped, RN21).
type InMatch struct {
	Room  string
	Event Event
}

// Released says these players are no longer at the table.
type Released struct {
	IDs []room.PlayerID
}

func (RoomsChanged) rummyEvent() {}
func (MatchStarted) rummyEvent() {}
func (InMatch) rummyEvent()      {}
func (Released) rummyEvent()     {}
