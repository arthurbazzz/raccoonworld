package cups

import (
	"time"

	"raccoons-world/server/internal/domain/room"
)

// Event is something that happened at the booth. The set is closed.
type Event interface{ isEvent() }

// StatusChanged says who is at the booth and what it is doing (RC6, RC22).
type StatusChanged struct {
	Phase     Phase
	Players   []room.PlayerID // in the lobby or the match
	Queued    []room.PlayerID // waiting for the next match
	LobbyLeft time.Duration   // time until the match starts; only in Lobby
}

// RoundStarted carries everything needed to animate a round (RC8). The swaps
// go out at once: anyone who reads them learns only what a player watching
// the cups learns (RC17).
type RoundStarted struct {
	Round    int
	Total    int
	Nut      int // slot the nut starts under
	Swaps    []Swap
	Show     time.Duration
	SwapTime time.Duration
	Pick     time.Duration
}

// Picked says a player picked, without the cup (RC10).
type Picked struct {
	ID room.PlayerID
}

// PickResult is one player's pick, revealed with the cups (RC8).
type PickResult struct {
	ID     room.PlayerID
	Slot   int // meaningless when Picked is false
	Picked bool
	Hit    bool
}

// Revealed lifts the cups: where the nut was and who hit.
type Revealed struct {
	Round int
	Nut   int
	Picks []PickResult
}

// Score is one line of the final scoreboard (RC12).
type Score struct {
	ID    room.PlayerID
	Hits  int
	Place int // ties share a place
}

// Finished is the end of a match, best first.
type Finished struct {
	Scores []Score
}

func (StatusChanged) isEvent() {}
func (RoundStarted) isEvent()  {}
func (Picked) isEvent()        {}
func (Revealed) isEvent()      {}
func (Finished) isEvent()      {}
