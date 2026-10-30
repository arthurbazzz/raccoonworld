package rummy_test

import (
	"slices"
	"testing"

	"raccoons-world/server/internal/domain/rummy"
)

func TestDeckHasEveryTileTwiceAndTwoJokers(t *testing.T) {
	deck := rummy.NewDeck()
	if len(deck) != 106 {
		t.Fatalf("deck has %d tiles, want 106 (RN1)", len(deck))
	}
	count := map[rummy.Tile]int{}
	jokers := 0
	for _, id := range deck {
		tile := rummy.TileOf(id)
		if tile.Joker {
			jokers++
			continue
		}
		if tile.Value < 1 || tile.Value > 13 || tile.Color >= rummy.Colors {
			t.Fatalf("TileOf(%d) = %+v is out of range", id, tile)
		}
		tile.ID = 0
		count[tile]++
	}
	if jokers != 2 {
		t.Fatalf("deck has %d jokers, want 2", jokers)
	}
	if len(count) != 52 {
		t.Fatalf("deck has %d distinct tiles, want 52", len(count))
	}
	for tile, n := range count {
		if n != 2 {
			t.Fatalf("%v appears %d times, want 2", tile, n)
		}
	}
}

func TestTileLayout(t *testing.T) {
	tests := []struct {
		id   rummy.TileID
		want string
	}{
		{0, "brown-1"}, {12, "brown-13"}, {13, "red-1"}, {51, "blue-13"},
		{52, "brown-1"}, {103, "blue-13"}, {104, "joker"}, {105, "joker"},
	}
	for _, tt := range tests {
		if got := rummy.TileOf(tt.id).String(); got != tt.want {
			t.Errorf("TileOf(%d) = %s, want %s", tt.id, got, tt.want)
		}
	}
	if rummy.ValidID(106) || !rummy.ValidID(105) {
		t.Fatal("ValidID must accept 0-105 only")
	}
}

func TestSeededShufflerRepeats(t *testing.T) {
	a, b := rummy.NewDeck(), rummy.NewDeck()
	rummy.NewRandomShuffler(7).Shuffle(a)
	rummy.NewRandomShuffler(7).Shuffle(b)
	if !slices.Equal(a, b) {
		t.Fatal("the same seed must give the same order")
	}
	if slices.Equal(a, rummy.NewDeck()) {
		t.Fatal("shuffled deck is still in order")
	}
}
