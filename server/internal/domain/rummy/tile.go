// Package rummy is the Rummy das Nozes minigame (RW-006): players take turns
// laying numbered tiles on a shared table, and whoever empties their hand
// first wins.
//
// Like the three cups, the game is pure: it never reads the clock nor draws
// random numbers by itself. Callers pass the time and a Shuffler, so tests
// control both. It is not safe for concurrent use; the room goroutine owns it.
package rummy

import (
	"fmt"
	"math/rand/v2"
)

// Color is one of the four tile colors, in the order of the art rows (Q6).
type Color uint8

// The tile colors.
const (
	Acorn  Color = iota // brown
	Cherry              // red
	Leaf                // green
	Drop                // blue
)

// Colors is how many colors there are.
const Colors = 4

// MaxValue is the highest tile number.
const MaxValue = 13

// DeckSize is how many tiles there are: 1 to 13 in 4 colors, twice, and 2
// jokers (RN1).
const DeckSize = 2*Colors*MaxValue + 2

// TileID identifies one physical tile, 0 to DeckSize-1. The two copies of a
// number have different IDs. The client knows what each ID is from the same
// fixed table (plan, section 3.5).
type TileID uint8

// Tile is what a TileID stands for.
type Tile struct {
	ID    TileID
	Color Color
	Value int // 1 to 13; 0 for a joker
	Joker bool
}

// TileOf returns the tile with the given ID. IDs are laid out as
// copy*52 + color*13 + (value-1); 104 and 105 are the jokers.
func TileOf(id TileID) Tile {
	n := int(id)
	if n >= 2*Colors*MaxValue {
		return Tile{ID: id, Joker: true}
	}
	n %= Colors * MaxValue
	return Tile{ID: id, Color: Color(n / MaxValue), Value: n%MaxValue + 1}
}

// ValidID reports whether id is a tile of the deck.
func ValidID(id TileID) bool { return int(id) < DeckSize }

func (t Tile) String() string {
	if t.Joker {
		return "joker"
	}
	return fmt.Sprintf("%s-%d", [...]string{"brown", "red", "green", "blue"}[t.Color], t.Value)
}

// NewDeck returns every tile ID, in order.
func NewDeck() []TileID {
	deck := make([]TileID, DeckSize)
	for i := range deck {
		deck[i] = TileID(i)
	}
	return deck
}

// Shuffler puts tiles in random order. Tests use a fixed order.
type Shuffler interface {
	Shuffle(tiles []TileID)
}

// RandomShuffler shuffles with math/rand: the order must be unpredictable,
// not secret, since it never leaves the server (RN24).
type RandomShuffler struct {
	rng *rand.Rand
}

// NewRandomShuffler returns a shuffler. A non-zero seed makes it repeat the
// same orders, for browser tests (RW_RUMMY_SEED); zero seeds it randomly.
func NewRandomShuffler(seed uint64) *RandomShuffler {
	if seed == 0 {
		seed = rand.Uint64()
	}
	return &RandomShuffler{rng: rand.New(rand.NewPCG(seed, seed))}
}

// Shuffle puts tiles in random order.
func (s *RandomShuffler) Shuffle(tiles []TileID) {
	s.rng.Shuffle(len(tiles), func(i, j int) { tiles[i], tiles[j] = tiles[j], tiles[i] })
}
