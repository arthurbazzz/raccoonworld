package usecase

import "raccoons-world/server/internal/domain/room"

// LeaveRoom removes a player whose connection closed and tells the room.
type LeaveRoom struct {
	runner      *RoomRunner
	broadcaster Broadcaster
}

// NewLeaveRoom builds the use case for one room.
func NewLeaveRoom(runner *RoomRunner, b Broadcaster) *LeaveRoom {
	return &LeaveRoom{runner: runner, broadcaster: b}
}

// Execute removes the player, from the three cups booth too (RC15). A seat
// in a rummy match is kept for a while (RW-006, RN21).
// Leaving twice, or after a timeout already removed them, is a no-op.
func (uc *LeaveRoom) Execute(id room.PlayerID) error {
	rr := uc.runner
	return rr.do(func() {
		if rr.room.Leave(id) {
			rr.leaveBooth(id, rr.clock.Now())
			rr.dropFromTable(id, rr.clock.Now())
			uc.broadcaster.ToRoom(rr.room.ID(), PlayerLeft{ID: id})
		}
	})
}

// KeepAlive records a sign of life from a player's connection (a heartbeat
// pong or any message). Players silent for room.PresenceTimeout are removed
// on the next tick (RN5).
type KeepAlive struct {
	runner *RoomRunner
	clock  Clock
}

// NewKeepAlive builds the use case for one room.
func NewKeepAlive(runner *RoomRunner, c Clock) *KeepAlive {
	return &KeepAlive{runner: runner, clock: c}
}

// Execute refreshes the player's presence. It returns
// room.ErrPlayerNotFound if they already left or timed out.
func (uc *KeepAlive) Execute(id room.PlayerID) error {
	now := uc.clock.Now()
	var touchErr error
	if err := uc.runner.Exec(func(r *room.Room) { touchErr = r.Touch(id, now) }); err != nil {
		return err
	}
	return touchErr
}
