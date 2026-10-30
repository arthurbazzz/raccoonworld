package rummy_test

import (
	"encoding/json"
	"os"
	"testing"

	"raccoons-world/server/internal/domain/rummy"
)

// meldCase is one row of testdata/melds.json, shared with the web client's
// tests so both sides agree on what a meld is (plan, section 6).
type meldCase struct {
	Name   string   `json:"name"`
	Tiles  []string `json:"tiles"`
	Kind   string   `json:"kind"`
	Points int      `json:"points"`
	Error  string   `json:"error"`
}

// idsByName maps "red-7" to a tile ID with that face, using a fresh copy for
// each repeat so two "red-7" are the two different red 7s.
func idsByName(t *testing.T, names []string) []rummy.TileID {
	t.Helper()
	used := map[rummy.TileID]bool{}
	ids := make([]rummy.TileID, 0, len(names))
	for _, name := range names {
		found := false
		for _, id := range rummy.NewDeck() {
			if !used[id] && rummy.TileOf(id).String() == name {
				used[id] = true
				ids = append(ids, id)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("no tile %q left", name)
		}
	}
	return ids
}

func TestClassify(t *testing.T) {
	data, err := os.ReadFile("testdata/melds.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []meldCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ids := idsByName(t, tc.Tiles)
			kind, err := rummy.Classify(ids)
			if tc.Error != "" {
				if err == nil || err.Error() != tc.Error {
					t.Fatalf("Classify = %v, %v; want error %s", kind, err, tc.Error)
				}
				if p := rummy.Points(ids); p != 0 {
					t.Fatalf("Points of an invalid meld = %d, want 0", p)
				}
				return
			}
			if err != nil || kind.String() != tc.Kind {
				t.Fatalf("Classify = %v, %v; want %s", kind, err, tc.Kind)
			}
			if p := rummy.Points(ids); p != tc.Points {
				t.Fatalf("Points = %d, want %d", p, tc.Points)
			}
		})
	}
}
