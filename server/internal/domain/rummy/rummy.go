package rummy

import "time"

// Table and match limits (spec RN5-RN8, RN13, RN18).
const (
	MinPlayers    = 2
	MaxPlayers    = 4
	HandSize      = 14
	OpeningPoints = 30 // minimum first play in the Classic mode
	JokerPenalty  = 30 // a joker left in the hand at the end
)

// Mode is chosen by whoever creates a room (RN13, Q2).
type Mode int

// The game modes.
const (
	Classic Mode = iota + 1 // first play needs 30 points from the hand
	Free                    // no minimum first play
)

func (m Mode) String() string {
	switch m {
	case Classic:
		return "classic"
	case Free:
		return "free"
	}
	return "unknown"
}

// ParseMode reads the wire name of a mode.
func ParseMode(s string) (Mode, bool) {
	switch s {
	case "classic":
		return Classic, true
	case "free":
		return Free, true
	}
	return 0, false
}

// Timing holds every duration of the table.
type Timing struct {
	Lobby   time.Duration // countdown once a room has 2 players (RN6)
	Turn    time.Duration // time to play a turn (RN15)
	Absence time.Duration // how long a dropped player keeps the seat (RN21)
	Results time.Duration // final scoreboard before the room reopens
}

// DefaultTiming is the real game speed (plan, section 3.1).
func DefaultTiming() Timing {
	return Timing{
		Lobby:   15 * time.Second,
		Turn:    40 * time.Second,
		Absence: 30 * time.Second,
		Results: 8 * time.Second,
	}
}

// FastTiming shortens only the lobby, for browser tests (RW_RUMMY_FAST). The
// turn keeps its 40 s, so there is time to drag tiles.
func FastTiming() Timing {
	t := DefaultTiming()
	t.Lobby = 1500 * time.Millisecond
	return t
}
