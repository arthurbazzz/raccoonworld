package room_test

import (
	"math"
	"testing"

	"raccoons-world/server/internal/domain/room"
)

const epsilon = 1e-9

func nearlyEqual(a, b float64) bool { return math.Abs(a-b) < epsilon }

func TestVecArithmetic(t *testing.T) {
	v := room.Vec{X: 3, Y: 4}

	if got := v.Len(); !nearlyEqual(got, 5) {
		t.Fatalf("Len() = %v, want 5", got)
	}
	if got := v.Add(room.Vec{X: 1, Y: -1}); got != (room.Vec{X: 4, Y: 3}) {
		t.Fatalf("Add() = %v, want {4 3}", got)
	}
	if got := v.Scale(2); got != (room.Vec{X: 6, Y: 8}) {
		t.Fatalf("Scale() = %v, want {6 8}", got)
	}
}

func TestVecNormalize(t *testing.T) {
	n := room.Vec{X: 3, Y: 4}.Normalize()
	if !nearlyEqual(n.Len(), 1) {
		t.Fatalf("Normalize().Len() = %v, want 1", n.Len())
	}
	if got := (room.Vec{}).Normalize(); got != (room.Vec{}) {
		t.Fatalf("zero.Normalize() = %v, want zero vector", got)
	}
}

func TestInputDirection(t *testing.T) {
	diag := 1 / math.Sqrt2

	// Screen coordinates: X grows right, Y grows down.
	tests := []struct {
		name  string
		input room.Input
		want  room.Vec
	}{
		{"idle", room.Input{}, room.Vec{}},
		{"up (W)", room.Input{Up: true}, room.Vec{X: 0, Y: -1}},
		{"left (A)", room.Input{Left: true}, room.Vec{X: -1, Y: 0}},
		{"down (S)", room.Input{Down: true}, room.Vec{X: 0, Y: 1}},
		{"right (D)", room.Input{Right: true}, room.Vec{X: 1, Y: 0}},
		{"up-right (W+D)", room.Input{Up: true, Right: true}, room.Vec{X: diag, Y: -diag}},
		{"down-left (S+A)", room.Input{Down: true, Left: true}, room.Vec{X: -diag, Y: diag}},
		{"opposites cancel (A+D)", room.Input{Left: true, Right: true}, room.Vec{}},
		{"opposites cancel (W+S)", room.Input{Up: true, Down: true}, room.Vec{}},
		{"three keys (W+A+D)", room.Input{Up: true, Left: true, Right: true}, room.Vec{X: 0, Y: -1}},
		{"all keys", room.Input{Up: true, Down: true, Left: true, Right: true}, room.Vec{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Direction()
			if !nearlyEqual(got.X, tt.want.X) || !nearlyEqual(got.Y, tt.want.Y) {
				t.Fatalf("Direction() = %v, want %v", got, tt.want)
			}
		})
	}
}

// RN3: diagonal speed must equal straight-line speed.
func TestInputDirectionHasUnitLengthWhenMoving(t *testing.T) {
	inputs := []room.Input{
		{Up: true}, {Right: true},
		{Up: true, Right: true}, {Down: true, Left: true},
	}
	for _, in := range inputs {
		if l := in.Direction().Len(); !nearlyEqual(l, 1) {
			t.Fatalf("%+v: Direction().Len() = %v, want 1", in, l)
		}
	}
}
