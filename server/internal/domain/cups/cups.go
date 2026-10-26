// Package cups is the three cups minigame (RW-005): a nut hides under one of
// three cups, the cups are shuffled and every player picks one.
//
// The game is a pure state machine. It never reads the clock nor draws random
// numbers by itself: callers pass the time and a Shuffler, so tests control
// both. It is not safe for concurrent use; the room goroutine owns it.
package cups

import (
	"math/rand/v2"
	"time"
)

// Slots is how many cups there are. Slots are numbered 0, 1, 2 from left to
// right; players see them as cups 1, 2 and 3.
const Slots = 3

// Swap exchanges whatever is in two slots.
type Swap [2]int

// RoundSpec is how hard one round is (Q4).
type RoundSpec struct {
	Swaps    int
	SwapTime time.Duration // how long each swap takes on screen
}

// Timing holds every duration of a match.
type Timing struct {
	Lobby   time.Duration // wait before a match starts (RC4)
	Show    time.Duration // the nut is shown before the shuffle (RC8)
	Pick    time.Duration // time to pick a cup (RC11)
	Reveal  time.Duration // cups stay up showing the result
	Results time.Duration // final scoreboard before the booth reopens
	Rounds  []RoundSpec   // one per round, each harder than the previous (RC7)
}

// DefaultTiming is the real game speed (plan, section 3.1).
func DefaultTiming() Timing {
	return Timing{
		Lobby:   10 * time.Second,
		Show:    1500 * time.Millisecond,
		Pick:    10 * time.Second,
		Reveal:  2500 * time.Millisecond,
		Results: 6 * time.Second,
		Rounds: []RoundSpec{
			{Swaps: 3, SwapTime: 600 * time.Millisecond},
			{Swaps: 5, SwapTime: 500 * time.Millisecond},
			{Swaps: 6, SwapTime: 400 * time.Millisecond},
			{Swaps: 8, SwapTime: 320 * time.Millisecond},
			{Swaps: 10, SwapTime: 250 * time.Millisecond},
		},
	}
}

// Scaled returns t with every duration divided by factor. Browser tests use
// it so a match does not take minutes (RW_CUPS_FAST).
func (t Timing) Scaled(factor int) Timing {
	if factor <= 1 {
		return t
	}
	d := func(v time.Duration) time.Duration { return v / time.Duration(factor) }
	out := Timing{
		Lobby: d(t.Lobby), Show: d(t.Show), Pick: d(t.Pick),
		Reveal: d(t.Reveal), Results: d(t.Results),
		Rounds: make([]RoundSpec, len(t.Rounds)),
	}
	for i, r := range t.Rounds {
		out.Rounds[i] = RoundSpec{Swaps: r.Swaps, SwapTime: d(r.SwapTime)}
	}
	return out
}

// Shuffler decides where the nut starts and which cups swap.
type Shuffler interface {
	// Start returns the slot the nut starts under.
	Start() int
	// Swaps returns n swaps, each between two different slots.
	Swaps(n int) []Swap
}

// RandomShuffler is the production Shuffler. The swaps are shown to every
// player, so they need to be unpredictable, not secret: math/rand is enough.
type RandomShuffler struct{}

// Start picks a random slot.
func (RandomShuffler) Start() int { return rand.IntN(Slots) }

// Swaps picks n random pairs of different slots.
func (RandomShuffler) Swaps(n int) []Swap {
	out := make([]Swap, n)
	for i := range out {
		a := rand.IntN(Slots)
		b := (a + 1 + rand.IntN(Slots-1)) % Slots
		out[i] = Swap{a, b}
	}
	return out
}

// FinalSlot follows the nut through the swaps (RC17): what a player who
// watched every swap would answer.
func FinalSlot(start int, swaps []Swap) int {
	slot := start
	for _, s := range swaps {
		switch slot {
		case s[0]:
			slot = s[1]
		case s[1]:
			slot = s[0]
		}
	}
	return slot
}
