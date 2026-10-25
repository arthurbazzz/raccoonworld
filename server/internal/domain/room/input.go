package room

// Input is the set of movement keys a player is holding (RN3):
// W = Up, A = Left, S = Down, D = Right.
type Input struct {
	Up, Down, Left, Right bool
}

// Direction returns the unit movement vector for the held keys. Opposite
// keys cancel out, and diagonals are normalized so they move at the same
// speed as straight lines.
func (in Input) Direction() Vec {
	var d Vec
	if in.Up {
		d.Y--
	}
	if in.Down {
		d.Y++
	}
	if in.Left {
		d.X--
	}
	if in.Right {
		d.X++
	}
	return d.Normalize()
}
