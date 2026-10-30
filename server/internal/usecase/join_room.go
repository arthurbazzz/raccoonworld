package usecase

import (
	"errors"
	"fmt"

	"raccoons-world/server/internal/domain/room"
)

// RejectReason tells a visitor why they could not join. The values are part
// of the WebSocket protocol (room.join_rejected).
type RejectReason string

// Join rejection reasons.
const (
	RejectNicknameTaken   RejectReason = "nickname_taken"
	RejectNicknameInvalid RejectReason = "nickname_invalid"
	RejectRoomFull        RejectReason = "room_full"
	// RejectAlreadyJoined: this connection is already in the room.
	RejectAlreadyJoined RejectReason = "already_joined"
)

// JoinRejectedError is returned by JoinRoom.Execute when the room refuses
// the visitor. The visitor has already been sent a JoinRejected event.
type JoinRejectedError struct {
	Reason RejectReason
}

func (e *JoinRejectedError) Error() string {
	return fmt.Sprintf("join rejected: %s", e.Reason)
}

// JoinInput is the request to enter a room. PlayerID is assigned by the
// transport adapter (one per connection).
type JoinInput struct {
	PlayerID room.PlayerID
	Nickname string
}

// JoinRoom puts a visitor in a room and tells everyone about it.
type JoinRoom struct {
	runner      *RoomRunner
	broadcaster Broadcaster
	clock       Clock
}

// NewJoinRoom builds the use case for one room.
func NewJoinRoom(runner *RoomRunner, b Broadcaster, c Clock) *JoinRoom {
	return &JoinRoom{runner: runner, broadcaster: b, clock: c}
}

// Execute validates the nickname and joins the room. On success the
// visitor gets Joined and the room gets PlayerJoined. On rejection the
// visitor gets JoinRejected and a *JoinRejectedError is returned.
func (uc *JoinRoom) Execute(in JoinInput) error {
	nickname, err := room.NewNickname(in.Nickname)
	if err != nil {
		return uc.reject(in.PlayerID, RejectNicknameInvalid)
	}

	now := uc.clock.Now()
	var joinErr error
	execErr := uc.runner.Exec(func(r *room.Room) {
		var me room.PlayerView
		me, joinErr = r.Join(in.PlayerID, nickname, now)
		if joinErr != nil {
			return
		}
		// Sent from the room goroutine so they are ordered with snapshots.
		uc.broadcaster.ToPlayer(in.PlayerID, Joined{
			RoomID:  r.ID(),
			MapName: r.Map().Name(),
			You:     me,
			Players: r.Players(),
			Cups:    uc.runner.cupsSnapshot(now),
			Rummy:   uc.runner.rummySnapshot(now),
		})
		uc.broadcaster.ToRoom(r.ID(), PlayerJoined{Player: me})
		uc.runner.reclaimSeat(me, now)
	})
	if execErr != nil {
		return execErr
	}

	switch {
	case joinErr == nil:
		return nil
	case errors.Is(joinErr, room.ErrNicknameTaken):
		return uc.reject(in.PlayerID, RejectNicknameTaken)
	case errors.Is(joinErr, room.ErrRoomFull):
		return uc.reject(in.PlayerID, RejectRoomFull)
	case errors.Is(joinErr, room.ErrPlayerExists):
		return uc.reject(in.PlayerID, RejectAlreadyJoined)
	default:
		return joinErr
	}
}

func (uc *JoinRoom) reject(id room.PlayerID, reason RejectReason) error {
	uc.broadcaster.ToPlayer(id, JoinRejected{Reason: reason})
	return &JoinRejectedError{Reason: reason}
}
