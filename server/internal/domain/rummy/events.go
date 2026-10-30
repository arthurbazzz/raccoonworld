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

// TurnEnded says how the turn ended. Reason explains a refused play (RN14); Drew
// is false when the pool was empty (RN16).
type TurnEnded struct {
	Turn   int
	By     room.PlayerID
	Result Result
	Reason string
	Drew   bool
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
func (PlayerRemoved) rummyEvent() {}
func (MatchOver) rummyEvent()     {}
