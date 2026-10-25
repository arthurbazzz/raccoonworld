// Package maps embeds the room tile maps so the server binary is
// self-contained. Layout of praca.json: border walls, four trees in the
// corners, four benches and a fountain in the center.
package maps

import _ "embed" // required for //go:embed

// Praca is the JSON tile map of the starting room (RW-001).
//
//go:embed praca.json
var Praca []byte
