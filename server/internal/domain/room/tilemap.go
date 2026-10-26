package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// ErrInvalidMap is returned when map data cannot be turned into a TileMap.
var ErrInvalidMap = errors.New("invalid map")

const (
	tileFloor = '.'
	tileSolid = '#'
	// tileBooth is a solid tile that is part of the three cups booth (RW-005).
	tileBooth = 'B'

	// edgeEpsilon keeps a box that ends exactly on a tile edge from counting
	// as overlapping the next tile.
	edgeEpsilon = 1e-9
)

// Rect is an axis-aligned box in room pixels; (X, Y) is its top-left corner.
type Rect struct {
	X, Y, W, H float64
}

// TileMap is the collision grid of a room. Anything outside the grid is
// solid, so players can never leave the room.
type TileMap struct {
	name     string
	tileSize float64
	cols     int
	rows     int
	solid    []bool // row-major, len = cols*rows
	spawn    Vec
	booth    *Rect // bounding box of the 'B' tiles; nil when the map has none
}

// BoothReach is how close, in pixels, a raccoon's feet must be to the booth
// to play (RW-005, RC2).
const BoothReach = 48.0

type tileMapJSON struct {
	Name     string   `json:"name"`
	TileSize float64  `json:"tileSize"`
	Spawn    *vecJSON `json:"spawn"`
	Rows     []string `json:"rows"`
}

type vecJSON struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ParseTileMap builds a TileMap from its JSON form. Each row is a string
// where '#' is a solid tile, 'B' a solid booth tile and '.' is floor.
func ParseTileMap(data []byte) (*TileMap, error) {
	var raw tileMapJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidMap, err)
	}
	if raw.TileSize <= 0 {
		return nil, fmt.Errorf("%w: tileSize must be positive", ErrInvalidMap)
	}
	if len(raw.Rows) == 0 || raw.Rows[0] == "" {
		return nil, fmt.Errorf("%w: map has no tiles", ErrInvalidMap)
	}

	cols := len(raw.Rows[0])
	m := &TileMap{
		name:     raw.Name,
		tileSize: raw.TileSize,
		cols:     cols,
		rows:     len(raw.Rows),
		solid:    make([]bool, 0, cols*len(raw.Rows)),
	}
	for r, row := range raw.Rows {
		if len(row) != cols {
			return nil, fmt.Errorf("%w: row %d has %d tiles, want %d", ErrInvalidMap, r, len(row), cols)
		}
		for c := 0; c < len(row); c++ {
			switch row[c] {
			case tileFloor:
				m.solid = append(m.solid, false)
			case tileSolid:
				m.solid = append(m.solid, true)
			case tileBooth:
				m.solid = append(m.solid, true)
				m.addBoothTile(c, r)
			default:
				return nil, fmt.Errorf("%w: unknown tile %q at row %d col %d", ErrInvalidMap, row[c], r, c)
			}
		}
	}

	if raw.Spawn == nil {
		return nil, fmt.Errorf("%w: missing spawn", ErrInvalidMap)
	}
	m.spawn = Vec{X: raw.Spawn.X, Y: raw.Spawn.Y}
	if m.IsSolid(m.cell(m.spawn.X), m.cell(m.spawn.Y)) {
		return nil, fmt.Errorf("%w: spawn %v is not on a floor tile", ErrInvalidMap, m.spawn)
	}
	return m, nil
}

func (m *TileMap) addBoothTile(col, row int) {
	t := Rect{X: float64(col) * m.tileSize, Y: float64(row) * m.tileSize, W: m.tileSize, H: m.tileSize}
	if m.booth == nil {
		m.booth = &t
		return
	}
	b := m.booth
	right, bottom := math.Max(b.X+b.W, t.X+t.W), math.Max(b.Y+b.H, t.Y+t.H)
	b.X, b.Y = math.Min(b.X, t.X), math.Min(b.Y, t.Y)
	b.W, b.H = right-b.X, bottom-b.Y
}

// Booth returns the box covered by the booth tiles, if the map has a booth.
func (m *TileMap) Booth() (Rect, bool) {
	if m.booth == nil {
		return Rect{}, false
	}
	return *m.booth, true
}

// NearBooth reports whether feet at pos are within BoothReach of the booth
// (RC2). The distance is from the point to the nearest edge of the box.
func (m *TileMap) NearBooth(pos Vec) bool {
	b, ok := m.Booth()
	if !ok {
		return false
	}
	dx := math.Max(0, math.Max(b.X-pos.X, pos.X-(b.X+b.W)))
	dy := math.Max(0, math.Max(b.Y-pos.Y, pos.Y-(b.Y+b.H)))
	return math.Hypot(dx, dy) <= BoothReach
}

// Spawn returns where new players appear, in room pixels.
func (m *TileMap) Spawn() Vec { return m.spawn }

// Name returns the map identifier (e.g. "praca").
func (m *TileMap) Name() string { return m.name }

// Cols returns the width in tiles.
func (m *TileMap) Cols() int { return m.cols }

// Rows returns the height in tiles.
func (m *TileMap) Rows() int { return m.rows }

// TileSize returns the side of a tile in pixels.
func (m *TileMap) TileSize() float64 { return m.tileSize }

// PixelSize returns the map width and height in pixels.
func (m *TileMap) PixelSize() (w, h float64) {
	return float64(m.cols) * m.tileSize, float64(m.rows) * m.tileSize
}

// IsSolid reports whether the tile at (col, row) blocks movement.
// Tiles outside the map are solid.
func (m *TileMap) IsSolid(col, row int) bool {
	if col < 0 || row < 0 || col >= m.cols || row >= m.rows {
		return true
	}
	return m.solid[row*m.cols+col]
}

// Move returns box moved by delta, stopped by solid tiles. Each axis is
// resolved separately so a box pushed diagonally into a wall slides along
// it. Large deltas are split into sub-steps of at most half a tile so the
// box cannot tunnel through thin walls.
func (m *TileMap) Move(box Rect, delta Vec) Rect {
	maxStep := m.tileSize / 2
	steps := int(math.Ceil(math.Max(math.Abs(delta.X), math.Abs(delta.Y)) / maxStep))
	if steps == 0 {
		return box
	}
	step := delta.Scale(1 / float64(steps))
	for i := 0; i < steps; i++ {
		box = m.moveX(box, step.X)
		box = m.moveY(box, step.Y)
	}
	return box
}

func (m *TileMap) moveX(b Rect, dx float64) Rect {
	if dx == 0 {
		return b
	}
	b.X += dx
	rowFirst, rowLast := m.span(b.Y, b.H)
	if dx > 0 {
		col := m.cell(b.X + b.W - edgeEpsilon)
		if m.anySolidInColumn(col, rowFirst, rowLast) {
			b.X = float64(col)*m.tileSize - b.W
		}
		return b
	}
	col := m.cell(b.X)
	if m.anySolidInColumn(col, rowFirst, rowLast) {
		b.X = float64(col+1) * m.tileSize
	}
	return b
}

func (m *TileMap) moveY(b Rect, dy float64) Rect {
	if dy == 0 {
		return b
	}
	b.Y += dy
	colFirst, colLast := m.span(b.X, b.W)
	if dy > 0 {
		row := m.cell(b.Y + b.H - edgeEpsilon)
		if m.anySolidInRow(row, colFirst, colLast) {
			b.Y = float64(row)*m.tileSize - b.H
		}
		return b
	}
	row := m.cell(b.Y)
	if m.anySolidInRow(row, colFirst, colLast) {
		b.Y = float64(row+1) * m.tileSize
	}
	return b
}

// cell converts a pixel coordinate into a tile index.
func (m *TileMap) cell(px float64) int {
	return int(math.Floor(px / m.tileSize))
}

// span returns the first and last tile indices covered by [start, start+size).
func (m *TileMap) span(start, size float64) (first, last int) {
	return m.cell(start), m.cell(start + size - edgeEpsilon)
}

func (m *TileMap) anySolidInColumn(col, rowFirst, rowLast int) bool {
	for r := rowFirst; r <= rowLast; r++ {
		if m.IsSolid(col, r) {
			return true
		}
	}
	return false
}

func (m *TileMap) anySolidInRow(row, colFirst, colLast int) bool {
	for c := colFirst; c <= colLast; c++ {
		if m.IsSolid(c, row) {
			return true
		}
	}
	return false
}
