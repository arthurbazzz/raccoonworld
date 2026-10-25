package usecase

import "raccoons-world/server/internal/domain/room"

// SetInputInput is the set of movement keys a player is holding (RN3).
// The client sends it only when the set changes.
type SetInputInput struct {
	PlayerID              room.PlayerID
	Up, Down, Left, Right bool
}

// SetInput records which keys a player holds. The room loop applies them on
// every tick until they change; the client never sends positions.
type SetInput struct {
	runner *RoomRunner
}

// NewSetInput builds the use case for one room.
func NewSetInput(runner *RoomRunner) *SetInput {
	return &SetInput{runner: runner}
}

// Execute stores the keys. It returns room.ErrPlayerNotFound if the player
// is not in the room.
func (uc *SetInput) Execute(in SetInputInput) error {
	keys := room.Input{Up: in.Up, Down: in.Down, Left: in.Left, Right: in.Right}
	var setErr error
	if err := uc.runner.Exec(func(r *room.Room) { setErr = r.SetInput(in.PlayerID, keys) }); err != nil {
		return err
	}
	return setErr
}
