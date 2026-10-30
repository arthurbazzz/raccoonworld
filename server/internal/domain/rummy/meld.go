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

// Issue explains why tiles are not a meld, or why a play is refused
// (RN12a): which tiles are to blame and the numbers a player needs to fix it.
// The client turns it into a sentence; the server only sends the code.
type Issue struct {
	Meld   int // index of the meld in the layout, or -1 for the whole play
	Code   string
	Tiles  []TileID       // the tiles to blame; empty means the whole meld
	Params map[string]int // color, value, expected, found, count, points, need
}

// Classify tells whether tiles, in the order they lie on the table, form a
// meld, and of which kind. A joker stands for whatever completes the meld
// (RN3). When tiles fit both kinds (one number and jokers), it is a run if
// their order allows, since that is how the player laid them out.
func Classify(ids []TileID) (MeldKind, error) {
	kind, issue := Diagnose(ids)
	if issue != nil {
		return 0, codeErr[issue.Code]
	}
	return kind, nil
}

var codeErr = map[string]error{}

func init() {
	for _, e := range []error{ErrTooShort, ErrTooLong, ErrRepeatedColor, ErrMixedColors, ErrGap, ErrOutOfRange} {
		codeErr[e.Error()] = e
	}
}

// Diagnose is Classify with the details of what is wrong (RN12a). Meld is
// always 0; callers set it.
func Diagnose(ids []TileID) (MeldKind, *Issue) {
	if len(ids) < MinMeld {
		return 0, &Issue{Code: ErrTooShort.Error(), Params: map[string]int{"count": len(ids)}}
	}
	tiles := tilesOf(ids)
	runIssue := checkRun(tiles)
	if runIssue == nil {
		return Run, nil
	}
	if !sameValue(tiles) {
		return 0, runIssue
	}
	if issue := checkGroup(tiles); issue != nil {
		return 0, issue
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

func checkGroup(tiles []Tile) *Issue {
	if len(tiles) > Colors {
		return &Issue{Code: ErrTooLong.Error(), Params: map[string]int{"count": len(tiles)}}
	}
	var seen [Colors]bool
	for _, t := range tiles {
		if t.Joker {
			continue
		}
		if seen[t.Color] {
			issue := &Issue{Code: ErrRepeatedColor.Error(), Params: map[string]int{"color": int(t.Color), "value": t.Value}}
			for _, o := range tiles {
				if !o.Joker && o.Color == t.Color {
					issue.Tiles = append(issue.Tiles, o.ID)
				}
			}
			return issue
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

func checkRun(tiles []Tile) *Issue {
	c := firstReal(tiles).Color
	var off []TileID
	for _, t := range tiles {
		if !t.Joker && t.Color != c {
			off = append(off, t.ID)
		}
	}
	if len(off) > 0 {
		return &Issue{Code: ErrMixedColors.Error(), Tiles: off, Params: map[string]int{"color": int(c)}}
	}
	start := runStart(tiles)
	for i, t := range tiles {
		if !t.Joker && t.Value != start+i {
			return &Issue{Code: ErrGap.Error(), Tiles: []TileID{t.ID}, Params: map[string]int{
				"color": int(c), "expected": start + i, "found": t.Value,
			}}
		}
	}
	if start < 1 || start+len(tiles)-1 > MaxValue {
		return &Issue{Code: ErrOutOfRange.Error(), Params: map[string]int{"start": start, "end": start + len(tiles) - 1}}
	}
	return nil
}
