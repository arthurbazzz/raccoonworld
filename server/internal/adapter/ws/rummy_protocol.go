package ws

import (
	"time"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/domain/rummy"
	"raccoons-world/server/internal/usecase"
)

// Rummy message types (see specs/006-rummy/plan.md, section 3.5). They are
// additions to protocol v1: older clients ignore unknown types.
const (
	TypeRummyOpen   = "rummy.open"
	TypeRummyCreate = "rummy.create"
	TypeRummyJoin   = "rummy.join"
	TypeRummyMove   = "rummy.move"
	TypeRummyEnd    = "rummy.end"
	TypeRummyDraw   = "rummy.draw"
	TypeRummyUndo   = "rummy.undo"
	TypeRummyLeave  = "rummy.leave"
	TypeRummyAgain  = "rummy.again"
	TypeRummySay    = "rummy.say"

	TypeRummyRooms    = "rummy.rooms"
	TypeRummyStatus   = "rummy.status"
	TypeRummyRejected = "rummy.rejected"
	TypeRummyState    = "rummy.state"
	TypeRummyDraft    = "rummy.draft"
	TypeRummyTurn     = "rummy.turn"
	TypeRummyResults  = "rummy.results"
	TypeRummyRefused  = "rummy.refused"
	TypeRummyChat     = "rummy.chat"
)

// RummyOpenRequest is "rummy.open": E near the table, to see the rooms.
type RummyOpenRequest struct{}

// RummyCreateRequest is "rummy.create": a new room in a mode.
type RummyCreateRequest struct {
	Mode string `json:"mode"`
}

// RummyJoinRequest is "rummy.join": enter a waiting room.
type RummyJoinRequest struct {
	Room string `json:"room"`
}

// RummyMoveRequest is "rummy.move": the whole draft after a drag. Tiles are
// IDs 0-105; anything else makes the layout invalid and it is ignored.
type RummyMoveRequest struct {
	Turn  int      `json:"turn"`
	Melds [][]int  `json:"melds"`
	Pos   [][2]int `json:"pos"` // [x, y] of each meld on the table grid (RN32)
	Hand  []int    `json:"hand"`
}

// RummySayRequest is "rummy.say": a chat line to the player's match (RN37).
type RummySayRequest struct {
	Text string `json:"text"`
}

func (RummySayRequest) isClientMessage() {}

// ToInput builds the SendChat input for the connection's player ID.
func (m RummySayRequest) ToInput(id room.PlayerID) usecase.SendChatInput {
	return usecase.SendChatInput{PlayerID: id, Text: m.Text, Table: true}
}

// RummyTurnRequest is "rummy.end", "rummy.draw" or "rummy.undo".
type RummyTurnRequest struct {
	Action usecase.TurnAction `json:"-"`
	Turn   int                `json:"turn"`
}

// RummyLeaveRequest is "rummy.leave": Esc or the Sair button.
type RummyLeaveRequest struct{}

// RummyAgainRequest is "rummy.again": Jogar de novo on the scoreboard.
type RummyAgainRequest struct{}

func (RummyOpenRequest) isClientMessage()   {}
func (RummyCreateRequest) isClientMessage() {}
func (RummyJoinRequest) isClientMessage()   {}
func (RummyMoveRequest) isClientMessage()   {}
func (RummyTurnRequest) isClientMessage()   {}
func (RummyLeaveRequest) isClientMessage()  {}
func (RummyAgainRequest) isClientMessage()  {}

// toTileIDs converts wire tiles. An ID out of range becomes one the layout
// check rejects, so the move is ignored.
func toTileIDs(in []int) []rummy.TileID {
	out := make([]rummy.TileID, len(in))
	for i, n := range in {
		if n < 0 || n >= rummy.DeckSize {
			n = rummy.DeckSize
		}
		out[i] = rummy.TileID(n)
	}
	return out
}

// ToInput builds the PlayTurn input for the connection's player ID.
func (m RummyMoveRequest) ToInput(id room.PlayerID) usecase.TurnInput {
	melds := make([][]rummy.TileID, len(m.Melds))
	for i, meld := range m.Melds {
		melds[i] = toTileIDs(meld)
	}
	in := usecase.TurnInput{PlayerID: id, Action: usecase.ActionMove, Turn: m.Turn, Melds: melds, Hand: toTileIDs(m.Hand)}
	if m.Pos != nil {
		in.Pos = make([]rummy.Pos, len(m.Pos))
		for i, p := range m.Pos {
			in.Pos[i] = rummy.Pos{X: p[0], Y: p[1]}
		}
	}
	return in
}

// posJSON encodes the positions of melds, laying them out when missing.
func posJSON(melds [][]rummy.TileID, pos []rummy.Pos) [][2]int {
	if len(pos) != len(melds) {
		pos = rummy.AutoPos(len(melds))
	}
	out := make([][2]int, len(pos))
	for i, p := range pos {
		out[i] = [2]int{p.X, p.Y}
	}
	return out
}

// ToInput builds the PlayTurn input for the connection's player ID.
func (m RummyTurnRequest) ToInput(id room.PlayerID) usecase.TurnInput {
	return usecase.TurnInput{PlayerID: id, Action: m.Action, Turn: m.Turn}
}

// decodeTurn decodes "rummy.end", "rummy.draw" and "rummy.undo".
func decodeTurn(a usecase.TurnAction, data []byte) (ClientMessage, error) {
	msg, err := decodeData[RummyTurnRequest](data)
	if err != nil {
		return nil, err
	}
	m, _ := msg.(RummyTurnRequest) // decodeData returns a RummyTurnRequest
	m.Action = a
	return m, nil
}

// Wire shapes of the rummy data.
type (
	rummyRoomJSON struct {
		ID      string          `json:"id"`
		Mode    string          `json:"mode"`
		Players []room.PlayerID `json:"players"`
		Phase   string          `json:"phase"`
		LobbyMs int64           `json:"lobbyMs"`
	}
	rummyRoomsJSON struct {
		Rooms []rummyRoomJSON `json:"rooms"`
	}
	rummyStatusJSON struct {
		Playing []room.PlayerID `json:"playing"`
		Tables  int             `json:"tables"`
	}
	rummySnapshotJSON struct {
		rummyStatusJSON
		Rooms []rummyRoomJSON `json:"rooms"`
	}
	rummyRejectedJSON struct {
		Reason usecase.RummyRejectReason `json:"reason"`
	}
	rummyStateJSON struct {
		Room    string                `json:"room"`
		Mode    string                `json:"mode"`
		Turn    int                   `json:"turn"`
		Current room.PlayerID         `json:"current"`
		TurnMs  int64                 `json:"turnMs"`
		Melds   [][]int               `json:"melds"`
		Pos     [][2]int              `json:"pos"`
		Hand    []int                 `json:"hand"`
		Counts  map[room.PlayerID]int `json:"counts"`
		Pool    int                   `json:"pool"`
		Opened  []room.PlayerID       `json:"opened"`
		Absent  []room.PlayerID       `json:"absent"`
	}
	rummyDraftJSON struct {
		Turn      int      `json:"turn"`
		Melds     [][]int  `json:"melds"`
		Pos       [][2]int `json:"pos"`
		HandCount int      `json:"handCount"`
	}
	rummyTurnJSON struct {
		Turn   int           `json:"turn"`
		Result rummy.Result  `json:"result"`
		By     room.PlayerID `json:"by"`
		Drew   bool          `json:"drew"`
	}
	rummyIssueJSON struct {
		Meld   int            `json:"meld"` // -1: the whole play
		Code   string         `json:"code"`
		Tiles  []int          `json:"tiles"`
		Params map[string]int `json:"params"`
	}
	rummyRefusedJSON struct {
		Turn   int              `json:"turn"`
		Issues []rummyIssueJSON `json:"issues"`
	}
	rummyScoreJSON struct {
		ID     room.PlayerID `json:"id"`
		Points int           `json:"points"`
		Place  int           `json:"place"`
	}
	rummyResultsJSON struct {
		Winner *room.PlayerID   `json:"winner"` // null on a tie
		Scores []rummyScoreJSON `json:"scores"`
	}
)

func tilesJSON(ids []rummy.TileID) []int {
	out := make([]int, len(ids))
	for i, id := range ids {
		out[i] = int(id)
	}
	return out
}

func meldsJSON(melds [][]rummy.TileID) [][]int {
	out := make([][]int, len(melds))
	for i, m := range melds {
		out[i] = tilesJSON(m)
	}
	return out
}

func ids(in []room.PlayerID) []room.PlayerID {
	if in == nil {
		return []room.PlayerID{}
	}
	return in
}

// left is the time from now until t, never negative.
func left(t, now time.Time) int64 {
	if t.IsZero() || !t.After(now) {
		return 0
	}
	return ms(t.Sub(now))
}

func toRummyRoomsJSON(r usecase.RummyRooms) []rummyRoomJSON {
	out := make([]rummyRoomJSON, len(r.Rooms))
	for i, info := range r.Rooms {
		players := make([]room.PlayerID, len(info.Players))
		for j, p := range info.Players {
			players[j] = p.ID
		}
		out[i] = rummyRoomJSON{
			ID: info.ID, Mode: info.Mode.String(), Players: players,
			Phase: info.Phase.String(), LobbyMs: left(info.LobbyEnds, r.Now),
		}
	}
	return out
}

func toRummyStatusJSON(s usecase.RummyStatus) rummyStatusJSON {
	return rummyStatusJSON{Playing: ids(s.Playing), Tables: s.Matches}
}

func toRummySnapshotJSON(s usecase.RummySnapshot) rummySnapshotJSON {
	return rummySnapshotJSON{rummyStatusJSON: toRummyStatusJSON(s.Status), Rooms: toRummyRoomsJSON(s.Rooms)}
}

func toRummyStateJSON(s usecase.RummyState) rummyStateJSON {
	v := s.View
	counts := v.Counts
	if counts == nil {
		counts = map[room.PlayerID]int{}
	}
	return rummyStateJSON{
		Room: s.Room, Mode: v.Mode.String(), Turn: v.Turn, Current: v.Current,
		TurnMs: left(v.TurnEnds, s.Now), Melds: meldsJSON(v.Table), Pos: posJSON(v.Table, v.Pos), Hand: tilesJSON(v.Hand),
		Counts: counts, Pool: v.Pool, Opened: ids(v.Opened), Absent: ids(v.Absent),
	}
}

func toRummyResultsJSON(r rummy.MatchOver) rummyResultsJSON {
	out := rummyResultsJSON{Scores: make([]rummyScoreJSON, len(r.Scores))}
	if r.Winner != "" {
		w := r.Winner
		out.Winner = &w
	}
	for i, s := range r.Scores {
		out.Scores[i] = rummyScoreJSON{ID: s.ID, Points: s.Points, Place: s.Place}
	}
	return out
}

// rummyWireData encodes the rummy events; ok is false for other events.
func rummyWireData(ev usecase.Event) (typ string, data any, ok bool) {
	switch e := ev.(type) {
	case usecase.RummyRooms:
		return TypeRummyRooms, rummyRoomsJSON{Rooms: toRummyRoomsJSON(e)}, true
	case usecase.RummyStatus:
		return TypeRummyStatus, toRummyStatusJSON(e), true
	case usecase.RummyRejected:
		return TypeRummyRejected, rummyRejectedJSON{Reason: e.Reason}, true
	case usecase.RummyState:
		return TypeRummyState, toRummyStateJSON(e), true
	case usecase.RummyDraft:
		return TypeRummyDraft, rummyDraftJSON{Turn: e.Turn, Melds: meldsJSON(e.Melds), Pos: posJSON(e.Melds, e.Pos), HandCount: e.HandCount}, true
	case usecase.RummyTurn:
		t := e.Turn
		return TypeRummyTurn, rummyTurnJSON{Turn: t.Turn, Result: t.Result, By: t.By, Drew: t.Drew}, true
	case usecase.RummyRefused:
		r := rummyRefusedJSON{Turn: e.Refused.Turn, Issues: make([]rummyIssueJSON, len(e.Refused.Issues))}
		for i, is := range e.Refused.Issues {
			params := is.Params
			if params == nil {
				params = map[string]int{}
			}
			r.Issues[i] = rummyIssueJSON{Meld: is.Meld, Code: is.Code, Tiles: tilesJSON(is.Tiles), Params: params}
		}
		return TypeRummyRefused, r, true
	case usecase.RummyChat:
		return TypeRummyChat, chatJSON{From: e.From, Text: e.Text}, true
	case usecase.RummyResults:
		return TypeRummyResults, toRummyResultsJSON(e.Results), true
	}
	return "", nil, false
}
