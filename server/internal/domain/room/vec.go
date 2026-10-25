package room

import "math"

// Vec is a 2D vector in room pixels. Screen coordinates: X grows to the
// right, Y grows downward.
type Vec struct {
	X, Y float64
}

// Add returns v + o.
func (v Vec) Add(o Vec) Vec { return Vec{X: v.X + o.X, Y: v.Y + o.Y} }

// Scale returns v multiplied by k.
func (v Vec) Scale(k float64) Vec { return Vec{X: v.X * k, Y: v.Y * k} }

// Len returns the Euclidean length of v.
func (v Vec) Len() float64 { return math.Hypot(v.X, v.Y) }

// Normalize returns v with length 1, or the zero vector when v is zero.
func (v Vec) Normalize() Vec {
	l := v.Len()
	if l == 0 {
		return Vec{}
	}
	return v.Scale(1 / l)
}
