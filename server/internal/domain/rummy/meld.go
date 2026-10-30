package rummy

import "errors"

// MeldKind is the type of a valid meld (RN2).
type MeldKind int

// The kinds of meld.
const (
	Group MeldKind = iota + 1 // same number, different colors
	Run                       // same color, consecutive numbers
)

func (k MeldKind) String() string {
	switch k {
	case Group:
		return "group"
	case Run:
		return "run"
	}
	return "invalid"
}

// Why a set of tiles is not a meld. The text is the reason code sent to the
// client (plan, section 3.5).
var (
	ErrTooShort      = errors.New("too_short")
	ErrTooLong       = errors.New("too_long")
	ErrRepeatedColor = errors.New("repeated_color")
	ErrMixedColors   = errors.New("mixed_colors")
	ErrGap           = errors.New("gap")
	ErrOutOfRange    = errors.New("out_of_range")
)

// MinMeld is the fewest tiles a meld can have (RN2).
const MinMeld = 3

// Classify tells whether tiles, in the order they lie on the table, form a
// meld, and of which kind. A joker stands for whatever completes the meld
// (RN3). When tiles fit both kinds (one number and jokers), it is a run if
// their order allows, since that is how the player laid them out.
func Classify(ids []TileID) (MeldKind, error) {
	if len(ids) < MinMeld {
		return 0, ErrTooShort
	}
	tiles := tilesOf(ids)
	runErr := checkRun(tiles)
	if runErr == nil {
		return Run, nil
	}
	if !sameValue(tiles) {
		return 0, runErr
	}
	if err := checkGroup(tiles); err != nil {
		return 0, err
	}
	return Group, nil
}

// Points is the sum of a valid meld's numbers, each joker counting as the
// number it stands for (RN13). It returns 0 for an invalid meld.
func Points(ids []TileID) int {
	kind, err := Classify(ids)
	if err != nil {
		return 0
	}
	tiles := tilesOf(ids)
	if kind == Group {
		return len(tiles) * firstReal(tiles).Value
	}
	start := runStart(tiles)
	return len(tiles)*start + len(tiles)*(len(tiles)-1)/2
}

func tilesOf(ids []TileID) []Tile {
	tiles := make([]Tile, len(ids))
	for i, id := range ids {
		tiles[i] = TileOf(id)
	}
	return tiles
}

// firstReal returns the first tile that is not a joker. A meld always has
// one, since there are only two jokers and a meld has at least three tiles.
func firstReal(tiles []Tile) Tile {
	for _, t := range tiles {
		if !t.Joker {
			return t
		}
	}
	return Tile{}
}

func sameValue(tiles []Tile) bool {
	v := firstReal(tiles).Value
	for _, t := range tiles {
		if !t.Joker && t.Value != v {
			return false
		}
	}
	return true
}

func checkGroup(tiles []Tile) error {
	if len(tiles) > Colors {
		return ErrTooLong
	}
	var seen [Colors]bool
	for _, t := range tiles {
		if t.Joker {
			continue
		}
		if seen[t.Color] {
			return ErrRepeatedColor
		}
		seen[t.Color] = true
	}
	return nil
}

// runStart is the number of the first position of a run, deduced from its
// first real tile.
func runStart(tiles []Tile) int {
	for i, t := range tiles {
		if !t.Joker {
			return t.Value - i
		}
	}
	return 0
}

func checkRun(tiles []Tile) error {
	c := firstReal(tiles).Color
	for _, t := range tiles {
		if !t.Joker && t.Color != c {
			return ErrMixedColors
		}
	}
	start := runStart(tiles)
	for i, t := range tiles {
		if !t.Joker && t.Value != start+i {
			return ErrGap
		}
	}
	if start < 1 || start+len(tiles)-1 > MaxValue {
		return ErrOutOfRange
	}
	return nil
}
