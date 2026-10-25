package room_test

import (
	"errors"
	"testing"

	"raccoons-world/server/internal/domain/room"
)

// testMap is 6x4 tiles of 32px. '#' is solid, '.' is floor.
//
//	col: 012345
//	row0 ######
//	row1 #..#.#
//	row2 #....#
//	row3 ######
const testMap = `{
  "name": "test",
  "tileSize": 32,
  "spawn": {"x": 64, "y": 80},
  "rows": [
    "######",
    "#..#.#",
    "#....#",
    "######"
  ]
}`

func mustParse(t *testing.T, data string) *room.TileMap {
	t.Helper()
	m, err := room.ParseTileMap([]byte(data))
	if err != nil {
		t.Fatalf("ParseTileMap: %v", err)
	}
	return m
}

func TestParseTileMap(t *testing.T) {
	m := mustParse(t, testMap)

	if m.Name() != "test" || m.Cols() != 6 || m.Rows() != 4 || m.TileSize() != 32 {
		t.Fatalf("got name=%q cols=%d rows=%d tile=%v", m.Name(), m.Cols(), m.Rows(), m.TileSize())
	}
	if m.Spawn() != (room.Vec{X: 64, Y: 80}) {
		t.Fatalf("Spawn() = %v, want {64 80}", m.Spawn())
	}
	if w, h := m.PixelSize(); w != 192 || h != 128 {
		t.Fatalf("PixelSize() = %v,%v, want 192,128", w, h)
	}

	solid := map[[2]int]bool{{0, 0}: true, {3, 1}: true, {1, 1}: false, {4, 2}: false}
	for cell, want := range solid {
		if got := m.IsSolid(cell[0], cell[1]); got != want {
			t.Fatalf("IsSolid(%d,%d) = %v, want %v", cell[0], cell[1], got, want)
		}
	}
}

func TestOutsideMapIsSolid(t *testing.T) {
	m := mustParse(t, testMap)
	for _, c := range [][2]int{{-1, 1}, {1, -1}, {6, 1}, {1, 4}} {
		if !m.IsSolid(c[0], c[1]) {
			t.Fatalf("IsSolid(%d,%d) = false, want true outside the map", c[0], c[1])
		}
	}
}

func TestParseTileMapErrors(t *testing.T) {
	tests := map[string]string{
		"not json":        `{`,
		"no rows":         `{"name":"x","tileSize":32,"spawn":{"x":48,"y":48},"rows":[]}`,
		"ragged rows":     `{"name":"x","tileSize":32,"spawn":{"x":48,"y":48},"rows":["###","##"]}`,
		"unknown tile":    `{"name":"x","tileSize":32,"spawn":{"x":48,"y":48},"rows":["#?#"]}`,
		"bad tile size":   `{"name":"x","tileSize":0,"spawn":{"x":48,"y":48},"rows":["###"]}`,
		"missing spawn":   `{"name":"x","tileSize":32,"rows":["###","#.#","###"]}`,
		"spawn on a wall": `{"name":"x","tileSize":32,"spawn":{"x":16,"y":16},"rows":["###","#.#","###"]}`,
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := room.ParseTileMap([]byte(data)); !errors.Is(err, room.ErrInvalidMap) {
				t.Fatalf("error = %v, want ErrInvalidMap", err)
			}
		})
	}
}

// hitbox is 20x12; a tile is 32px.
func box(x, y float64) room.Rect { return room.Rect{X: x, Y: y, W: 20, H: 12} }

func TestMoveFreely(t *testing.T) {
	m := mustParse(t, testMap)
	got := m.Move(box(40, 70), room.Vec{X: 10, Y: 5})
	if got != box(50, 75) {
		t.Fatalf("Move() = %+v, want %+v", got, box(50, 75))
	}
}

func TestMoveStopsFlushAgainstWall(t *testing.T) {
	m := mustParse(t, testMap)

	// Row 1: tile (3,1) is solid, its left edge is x=96. Box right edge must stop there.
	got := m.Move(box(70, 40), room.Vec{X: 20})
	if got.X+got.W != 96 {
		t.Fatalf("right edge = %v, want 96 (flush with wall)", got.X+got.W)
	}

	// Left border wall: tile column 0 ends at x=32.
	got = m.Move(box(40, 70), room.Vec{X: -30})
	if got.X != 32 {
		t.Fatalf("X = %v, want 32 (flush with left border)", got.X)
	}
}

func TestMoveSlidesAlongWall(t *testing.T) {
	m := mustParse(t, testMap)

	// Pushing up-left against the left border: X is blocked, Y keeps moving.
	got := m.Move(box(32, 80), room.Vec{X: -8, Y: -8})
	if got.X != 32 || got.Y != 72 {
		t.Fatalf("Move() = %+v, want X=32 Y=72 (slide)", got)
	}
}

func TestMoveNeverLeavesMap(t *testing.T) {
	m := mustParse(t, testMap)
	got := m.Move(box(40, 70), room.Vec{X: -1000, Y: 1000})
	if got.X < 32 || got.Y+got.H > 96 {
		t.Fatalf("Move() = %+v escaped the walkable area", got)
	}
}

func TestMoveDoesNotTunnelThroughWalls(t *testing.T) {
	m := mustParse(t, testMap)
	// A huge step right from column 1 must stop at the wall in column 3.
	got := m.Move(box(40, 40), room.Vec{X: 500})
	if got.X+got.W > 96 {
		t.Fatalf("Move() = %+v tunneled through the wall at x=96", got)
	}
}
