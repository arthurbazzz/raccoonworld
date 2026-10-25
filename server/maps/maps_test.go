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
