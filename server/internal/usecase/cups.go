package usecase

import "raccoons-world/server/internal/domain/room"

// JoinCups puts a player at the three cups booth: in the lobby, or queued
// for the next match if one is on (RC4, RC5). The server checks the raccoon
// is near the booth, using its own position (RC2).
type JoinCups struct {
	runner      *RoomRunner
	broadcaster Broadcaster
	clock       Clock
}

// NewJoinCups builds the use case for one room.
func NewJoinCups(runner *RoomRunner, b Broadcaster, c Clock) *JoinCups {
	return &JoinCups{runner: runner, broadcaster: b, clock: c}
}

// Execute joins the booth. A raccoon too far away gets CupsRejected. It
// returns room.ErrPlayerNotFound if the player is not in the room.
func (uc *JoinCups) Execute(id room.PlayerID) error {
	rr := uc.runner
	var err error
	if execErr := rr.do(func() {
		p, ok := rr.room.Player(id)
		if !ok {
			err = room.ErrPlayerNotFound
			return
		}
		if !rr.room.Map().NearBooth(p.Position) {
			uc.broadcaster.ToPlayer(id, CupsRejected{Reason: CupsTooFar})
			return
		}
		if rr.rummyBusy[id] {
			uc.broadcaster.ToPlayer(id, CupsRejected{Reason: CupsBusy})
			return
		}
		_, events := rr.game.Join(id, uc.clock.Now())
		rr.publishCups(events)
	}); execErr != nil {
		return execErr
	}
	return err
}

// PickInput is a player's pick: the cup slot (0-2, left to right) for a
// round. The round number keeps a late pick from landing in the next round.
type PickInput struct {
	PlayerID room.PlayerID
	Round    int
	Slot     int
}

// PickCup records a pick. Picks at the wrong time, repeated or from someone
// not in the match are ignored, like the rate-limited chat (RC9, RC18).
type PickCup struct {
	runner *RoomRunner
	clock  Clock
}

// NewPickCup builds the use case for one room.
func NewPickCup(runner *RoomRunner, c Clock) *PickCup {
	return &PickCup{runner: runner, clock: c}
}

// Execute records the pick, if it is valid.
func (uc *PickCup) Execute(in PickInput) error {
	rr := uc.runner
	return rr.do(func() {
		events, _ := rr.game.Pick(in.PlayerID, in.Round, in.Slot, uc.clock.Now())
		rr.publishCups(events)
	})
}

// LeaveCups takes a player out of the booth; the match goes on for the
// others (RC13, RC14).
type LeaveCups struct {
	runner *RoomRunner
	clock  Clock
}

// NewLeaveCups builds the use case for one room.
func NewLeaveCups(runner *RoomRunner, c Clock) *LeaveCups {
	return &LeaveCups{runner: runner, clock: c}
}

// Execute leaves the booth. Leaving when not there is a no-op.
func (uc *LeaveCups) Execute(id room.PlayerID) error {
	rr := uc.runner
	return rr.do(func() { rr.leaveBooth(id, uc.clock.Now()) })
}
