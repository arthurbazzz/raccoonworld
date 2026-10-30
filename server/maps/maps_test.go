package maps_test

import (
	"testing"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/maps"
)

func TestPracaMap(t *testing.T) {
	m, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		t.Fatalf("ParseTileMap(praca): %v", err)
	}
	if m.Name() != "praca" || m.Cols() != 30 || m.Rows() != 20 || m.TileSize() != 32 {
		t.Fatalf("got name=%q %dx%d tile=%v, want praca 30x20 tile=32",
			m.Name(), m.Cols(), m.Rows(), m.TileSize())
	}

	for c := 0; c < m.Cols(); c++ {
		if !m.IsSolid(c, 0) || !m.IsSolid(c, m.Rows()-1) {
			t.Fatalf("border column %d is not solid", c)
		}
	}
	for r := 0; r < m.Rows(); r++ {
		if !m.IsSolid(0, r) || !m.IsSolid(m.Cols()-1, r) {
			t.Fatalf("border row %d is not solid", r)
		}
	}
}

// The three cups booth sits above the fountain, away from the spawn point
// (RW-005, plan section 2).
func TestPracaBooth(t *testing.T) {
	m, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := m.Booth()
	if want := (room.Rect{X: 448, Y: 128, W: 96, H: 32}); !ok || b != want {
		t.Fatalf("Praça booth = %+v (%v), want %+v", b, ok, want)
	}
	if m.NearBooth(m.Spawn()) {
		t.Fatal("the spawn point must be away from the booth")
	}
}
