package usecase

import (
	"time"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/domain/rummy"
)

// RummyRooms is the room list of the pieces table (RN5-RN7), broadcast to
// the whole room whenever it changes and sent to whoever opens it.
type RummyRooms struct {
	Rooms []rummy.RoomInfo
	Now   time.Time // to turn the countdowns into durations
}

// RummyStatus is who is at the table and how many rooms are playing, for
// the "jogando" tag and the table in the Praça (RN25, RN27).
type RummyStatus struct {
	Playing []room.PlayerID
	Matches int
}

// RummyRejectReason tells a player why the table refused them.
type RummyRejectReason string

// Rejection reasons of the table. The domain errors add full, playing,
// busy, no_room and bad_mode.
const (
	RummyTooFar RummyRejectReason = "too_far"
	RummyAtCups RummyRejectReason = "busy" // playing the three cups
)

// RummyRejected is sent to a player whose table request was refused.
type RummyRejected struct {
	Reason RummyRejectReason
}

// RummyState is one player's view of their match. It is the only event with
// a hand, and each player gets their own (RN9, RN24).
type RummyState struct {
	Room string
	View rummy.View
	Now  time.Time // to turn the turn deadline into a duration
}

// RummyDraft is the current player's arrangement, sent live to the others in
// the match (plan, section 3.5).
type RummyDraft struct {
	Turn      int
	Melds     [][]rummy.TileID
	HandCount int
}

// RummyTurn is how a turn ended, sent to the match.
type RummyTurn struct {
	Turn rummy.TurnEnded
}

// RummyResults is the scoreboard at the end of a match (RN19).
type RummyResults struct {
	Results rummy.MatchOver
}

// RummySnapshot is the table as a newcomer to the room must see it.
type RummySnapshot struct {
	Status RummyStatus
	Rooms  RummyRooms
}

func (RummyRooms) isEvent()    {}
func (RummyStatus) isEvent()   {}
func (RummyRejected) isEvent() {}
func (RummyState) isEvent()    {}
func (RummyDraft) isEvent()    {}
func (RummyTurn) isEvent()     {}
func (RummyResults) isEvent()  {}

func (rr *RoomRunner) rummyStatus() RummyStatus {
	return RummyStatus{Playing: rr.table.Playing(), Matches: rr.table.Matches()}
}

func (rr *RoomRunner) rummySnapshot(now time.Time) RummySnapshot {
	return RummySnapshot{Status: rr.rummyStatus(), Rooms: RummyRooms{Rooms: rr.table.Rooms(), Now: now}}
}

// publishRummy turns table events into messages. Each match player gets
// their own state; the room gets the list and the status.
func (rr *RoomRunner) publishRummy(events []rummy.Event, now time.Time) {
	if len(events) == 0 {
		return
	}
	roomID := rr.room.ID()
	dirty := map[string]bool{} // match rooms whose players need a new state
	listChanged := false
	for _, e := range events {
		switch e := e.(type) {
		case rummy.RoomsChanged:
			listChanged = true
		case rummy.MatchStarted:
			dirty[e.Room] = true
		case rummy.InMatch:
			dirty[e.Room] = true
			switch me := e.Event.(type) {
			case rummy.TurnEnded:
				rr.toMatch(e.Room, RummyTurn{Turn: me})
			case rummy.MatchOver:
				for _, s := range me.Scores {
					rr.broadcaster.ToPlayer(s.ID, RummyResults{Results: me})
				}
			}
		}
	}
	for r := range dirty {
		rr.sendStates(r, now)
	}
	if listChanged {
		rr.broadcaster.ToRoom(roomID, RummyRooms{Rooms: rr.table.Rooms(), Now: now})
	}
	rr.broadcaster.ToRoom(roomID, rr.rummyStatus())
	rr.syncRummyBusy()
}

func (rr *RoomRunner) toMatch(roomID string, ev Event) {
	if m, ok := rr.table.Match(roomID); ok {
		for _, p := range m.Players() {
			rr.broadcaster.ToPlayer(p.ID, ev)
		}
	}
}

// sendStates sends each player of a match their own view of it.
func (rr *RoomRunner) sendStates(roomID string, now time.Time) {
	m, ok := rr.table.Match(roomID)
	if !ok {
		return
	}
	if _, over := m.Over(); over {
		return
	}
	for _, p := range m.Players() {
		rr.broadcaster.ToPlayer(p.ID, RummyState{Room: roomID, View: m.ViewFor(p.ID), Now: now})
	}
}

// sendDraft shows the current player's arrangement to the rest of the match.
func (rr *RoomRunner) sendDraft(roomID string, from room.PlayerID) {
	m, ok := rr.table.Match(roomID)
	if !ok {
		return
	}
	turn, melds, n := m.Draft()
	for _, p := range m.Players() {
		if p.ID != from {
			rr.broadcaster.ToPlayer(p.ID, RummyDraft{Turn: turn, Melds: melds, HandCount: n})
		}
	}
}

func (rr *RoomRunner) syncRummyBusy() {
	now := map[room.PlayerID]bool{}
	for _, p := range rr.table.Playing() {
		now[p] = true
	}
	before := rr.rummyBusy
	rr.rummyBusy = now
	rr.applyBusy(before, now)
}

// dropFromTable handles a player who left the Praça: a match keeps the seat
// for a while (RN21); a lobby or scoreboard simply lets them go.
func (rr *RoomRunner) dropFromTable(id room.PlayerID, now time.Time) {
	rr.publishRummy(rr.table.SetAbsent(id, now), now)
}

// reclaimSeat gives a returning player their seat back (plan, section 3.2).
func (rr *RoomRunner) reclaimSeat(p room.PlayerView, now time.Time) {
	if roomID, _, ok := rr.table.Reclaim(p.Nickname, p.ID); ok {
		rr.publishRummy([]rummy.Event{rummy.InMatch{Room: roomID}}, now)
	}
}

// tableRequest runs fn for a player near the table who is not at the booth.
// It returns room.ErrPlayerNotFound if the player is not in the room.
func (rr *RoomRunner) tableRequest(id room.PlayerID, fn func(p room.PlayerView)) error {
	var err error
	if execErr := rr.do(func() {
		p, ok := rr.room.Player(id)
		switch {
		case !ok:
			err = room.ErrPlayerNotFound
		case !rr.room.Map().NearTable(p.Position):
			rr.broadcaster.ToPlayer(id, RummyRejected{Reason: RummyTooFar})
		case rr.cupsBusy[id]:
			rr.broadcaster.ToPlayer(id, RummyRejected{Reason: RummyAtCups})
		default:
			fn(p)
		}
	}); execErr != nil {
		return execErr
	}
	return err
}

func (rr *RoomRunner) rejectRummy(id room.PlayerID, err error) {
	rr.broadcaster.ToPlayer(id, RummyRejected{Reason: RummyRejectReason(err.Error())})
}

// OpenTable sends the room list to a player near the table (E, RN4, RN5).
type OpenTable struct {
	runner *RoomRunner
	clock  Clock
}

// NewOpenTable builds the use case for one room.
func NewOpenTable(runner *RoomRunner, c Clock) *OpenTable {
	return &OpenTable{runner: runner, clock: c}
}

// Execute sends the list, or RummyRejected when too far.
func (uc *OpenTable) Execute(id room.PlayerID) error {
	rr := uc.runner
	return rr.tableRequest(id, func(room.PlayerView) {
		rr.broadcaster.ToPlayer(id, RummyRooms{Rooms: rr.table.Rooms(), Now: uc.clock.Now()})
	})
}

// CreateTableRoom opens a room at the table in a mode (RN5, RN13).
type CreateTableRoom struct {
	runner *RoomRunner
	clock  Clock
}

// NewCreateTableRoom builds the use case for one room.
func NewCreateTableRoom(runner *RoomRunner, c Clock) *CreateTableRoom {
	return &CreateTableRoom{runner: runner, clock: c}
}

// Execute creates the room. Refusals reach the player as RummyRejected.
func (uc *CreateTableRoom) Execute(id room.PlayerID, mode string) error {
	rr := uc.runner
	return rr.tableRequest(id, func(p room.PlayerView) {
		m, ok := rummy.ParseMode(mode)
		if !ok {
			rr.rejectRummy(id, rummy.ErrNoMode)
			return
		}
		_, events, err := rr.table.Create(rummy.Player{ID: p.ID, Nickname: p.Nickname}, m)
		if err != nil {
			rr.rejectRummy(id, err)
			return
		}
		rr.publishRummy(events, uc.clock.Now())
	})
}

// JoinTableRoom puts a player in a waiting room (RN5-RN7).
type JoinTableRoom struct {
	runner *RoomRunner
	clock  Clock
}

// NewJoinTableRoom builds the use case for one room.
func NewJoinTableRoom(runner *RoomRunner, c Clock) *JoinTableRoom {
	return &JoinTableRoom{runner: runner, clock: c}
}

// Execute joins the room. Refusals reach the player as RummyRejected.
func (uc *JoinTableRoom) Execute(id room.PlayerID, roomID string) error {
	rr := uc.runner
	return rr.tableRequest(id, func(p room.PlayerView) {
		events, err := rr.table.Join(rummy.Player{ID: p.ID, Nickname: p.Nickname}, roomID, uc.clock.Now())
		if err != nil {
			rr.rejectRummy(id, err)
			return
		}
		rr.publishRummy(events, uc.clock.Now())
	})
}

// LeaveTableRoom takes a player out of the table (RN20). Leaving when not
// there is a no-op.
type LeaveTableRoom struct {
	runner *RoomRunner
	clock  Clock
}

// NewLeaveTableRoom builds the use case for one room.
func NewLeaveTableRoom(runner *RoomRunner, c Clock) *LeaveTableRoom {
	return &LeaveTableRoom{runner: runner, clock: c}
}

// Execute leaves the table.
func (uc *LeaveTableRoom) Execute(id room.PlayerID) error {
	rr := uc.runner
	return rr.do(func() {
		now := uc.clock.Now()
		rr.publishRummy(rr.table.Leave(id, now), now)
	})
}

// PlayAgain keeps a player in their room for the next match, from the
// scoreboard (RN7, RN19).
type PlayAgain struct {
	runner *RoomRunner
}

// NewPlayAgain builds the use case for one room.
func NewPlayAgain(runner *RoomRunner) *PlayAgain {
	return &PlayAgain{runner: runner}
}

// Execute marks the player.
func (uc *PlayAgain) Execute(id room.PlayerID) error {
	rr := uc.runner
	return rr.do(func() { rr.table.PlayAgain(id) })
}

// TurnAction is what a player does on their turn.
type TurnAction int

// The turn actions (plan, section 3.5).
const (
	ActionMove TurnAction = iota + 1 // rearrange the draft
	ActionEnd                        // end the turn (RN14)
	ActionDraw                       // draw a tile (RN10)
	ActionUndo                       // back to the start of the turn (RN12)
)

// TurnInput is one turn action. Melds and Hand are only used by ActionMove.
// Turn keeps a late action from landing on the next turn.
type TurnInput struct {
	PlayerID room.PlayerID
	Action   TurnAction
	Turn     int
	Melds    [][]rummy.TileID
	Hand     []rummy.TileID
}

// PlayTurn applies a turn action. Actions out of turn, late or with a
// layout that does not hold the same tiles are ignored: the client gets its
// state again, so it snaps back to what the server has.
type PlayTurn struct {
	runner *RoomRunner
	clock  Clock
}

// NewPlayTurn builds the use case for one room.
func NewPlayTurn(runner *RoomRunner, c Clock) *PlayTurn {
	return &PlayTurn{runner: runner, clock: c}
}

// Execute applies the action.
func (uc *PlayTurn) Execute(in TurnInput) error {
	rr := uc.runner
	return rr.do(func() {
		roomID, ok := rr.table.RoomOf(in.PlayerID)
		if !ok {
			return
		}
		m, ok := rr.table.Match(roomID)
		if !ok {
			return
		}
		now := uc.clock.Now()
		var events []rummy.Event
		var err error
		switch in.Action {
		case ActionMove:
			if err = m.Move(in.PlayerID, in.Turn, rummy.Layout{Melds: in.Melds, Hand: in.Hand}); err == nil {
				rr.sendDraft(roomID, in.PlayerID)
				return
			}
		case ActionUndo:
			if err = m.Undo(in.PlayerID, in.Turn); err == nil {
				rr.sendDraft(roomID, in.PlayerID)
			}
		case ActionEnd:
			events, err = m.EndTurn(in.PlayerID, in.Turn, now)
		case ActionDraw:
			events, err = m.Draw(in.PlayerID, in.Turn, now)
		}
		if err != nil && m.Has(in.PlayerID) {
			if _, over := m.Over(); !over {
				rr.broadcaster.ToPlayer(in.PlayerID, RummyState{Room: roomID, View: m.ViewFor(in.PlayerID), Now: now})
			}
			return
		}
		if in.Action == ActionUndo {
			rr.broadcaster.ToPlayer(in.PlayerID, RummyState{Room: roomID, View: m.ViewFor(in.PlayerID), Now: now})
			return
		}
		rr.publishRummy(rr.table.Settle(roomID, events, now), now)
	})
}
