// Package ws is the WebSocket adapter: it translates protocol v1 JSON
// messages into use case calls and use case events into JSON.
package ws

import (
	"encoding/json"
	"errors"
	"fmt"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// ProtocolVersion is the only version this server speaks.
const ProtocolVersion = 1

// Message types (see specs/001-sala-inicial/plan.md, section 3.4).
const (
	TypeRoomJoin = "room.join"
	TypeInput    = "player.input"
	TypeChatSend = "chat.send"

	TypeRoomJoined       = "room.joined"
	TypeRoomJoinRejected = "room.join_rejected"
	TypePlayerJoined     = "player.joined"
	TypePlayerLeft       = "player.left"
	TypeSnapshot         = "room.snapshot"
	TypeChatMessage      = "chat.message"
)

var (
	// ErrMalformed means the message is not a valid protocol envelope.
	ErrMalformed = errors.New("malformed message")
	// ErrUnsupportedVersion means "v" is missing or not ProtocolVersion.
	ErrUnsupportedVersion = errors.New("unsupported protocol version")
	// ErrUnknownType means the type is not a client-to-server message.
	ErrUnknownType = errors.New("unknown message type")
)

type envelope struct {
	Type string          `json:"type"`
	V    int             `json:"v"`
	Data json.RawMessage `json:"data"`
}

// ClientMessage is a decoded client-to-server message. The set is closed:
// JoinRequest, InputRequest, ChatRequest or one of the Cups* and Rummy*
// requests.
type ClientMessage interface {
	isClientMessage()
}

// JoinRequest is "room.join".
type JoinRequest struct {
	Nickname string `json:"nickname"`
}

// InputRequest is "player.input": the movement keys held right now.
// Seq is echoed for future client-side reconciliation and unused today.
type InputRequest struct {
	Up    bool `json:"up"`
	Down  bool `json:"down"`
	Left  bool `json:"left"`
	Right bool `json:"right"`
	Seq   int  `json:"seq"`
}

// ChatRequest is "chat.send".
type ChatRequest struct {
	Text string `json:"text"`
}

func (JoinRequest) isClientMessage()  {}
func (InputRequest) isClientMessage() {}
func (ChatRequest) isClientMessage()  {}

// ToInput builds the JoinRoom input for the connection's player ID.
func (m JoinRequest) ToInput(id room.PlayerID) usecase.JoinInput {
	return usecase.JoinInput{PlayerID: id, Nickname: m.Nickname}
}

// ToInput builds the SetInput input for the connection's player ID.
func (m InputRequest) ToInput(id room.PlayerID) usecase.SetInputInput {
	return usecase.SetInputInput{PlayerID: id, Up: m.Up, Down: m.Down, Left: m.Left, Right: m.Right}
}

// ToInput builds the SendChat input for the connection's player ID.
func (m ChatRequest) ToInput(id room.PlayerID) usecase.SendChatInput {
	return usecase.SendChatInput{PlayerID: id, Text: m.Text}
}

// DecodeClient parses one client-to-server message. Unknown fields are
// ignored so newer clients can add optional data.
func DecodeClient(raw []byte) (ClientMessage, error) {
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	if env.V != ProtocolVersion {
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedVersion, env.V)
	}

	switch env.Type {
	case TypeRoomJoin:
		return decodeData[JoinRequest](env.Data)
	case TypeInput:
		return decodeData[InputRequest](env.Data)
	case TypeChatSend:
		return decodeData[ChatRequest](env.Data)
	case TypeCupsJoin:
		return decodeData[CupsJoinRequest](env.Data)
	case TypeCupsPick:
		return decodeData[CupsPickRequest](env.Data)
	case TypeCupsLeave:
		return decodeData[CupsLeaveRequest](env.Data)
	case TypeRummyOpen:
		return decodeData[RummyOpenRequest](env.Data)
	case TypeRummyCreate:
		return decodeData[RummyCreateRequest](env.Data)
	case TypeRummyJoin:
		return decodeData[RummyJoinRequest](env.Data)
	case TypeRummyMove:
		return decodeData[RummyMoveRequest](env.Data)
	case TypeRummyEnd:
		return decodeTurn(usecase.ActionEnd, env.Data)
	case TypeRummyDraw:
		return decodeTurn(usecase.ActionDraw, env.Data)
	case TypeRummyUndo:
		return decodeTurn(usecase.ActionUndo, env.Data)
	case TypeRummyLeave:
		return decodeData[RummyLeaveRequest](env.Data)
	case TypeRummyAgain:
		return decodeData[RummyAgainRequest](env.Data)
	case TypeRummySay:
		return decodeData[RummySayRequest](env.Data)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownType, env.Type)
	}
}

// decodeData unmarshals the "data" object into the request type T.
func decodeData[T ClientMessage](data json.RawMessage) (ClientMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: missing data", ErrMalformed)
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	return v, nil
}

// Wire shapes of server-to-client data.
type (
	playerJSON struct {
		ID       room.PlayerID `json:"id"`
		Nickname string        `json:"nickname"`
		X        float64       `json:"x"`
		Y        float64       `json:"y"`
	}
	positionJSON struct {
		ID room.PlayerID `json:"id"`
		X  float64       `json:"x"`
		Y  float64       `json:"y"`
	}
	joinedJSON struct {
		You     room.PlayerID     `json:"you"`
		Map     string            `json:"map"`
		Players []playerJSON      `json:"players"`
		Cups    cupsSnapshotJSON  `json:"cups"`
		Rummy   rummySnapshotJSON `json:"rummy"`
	}
	rejectedJSON struct {
		Reason usecase.RejectReason `json:"reason"`
	}
	leftJSON struct {
		ID room.PlayerID `json:"id"`
	}
	snapshotJSON struct {
		Tick    uint64         `json:"tick"`
		Players []positionJSON `json:"players"`
	}
	chatJSON struct {
		From room.PlayerID `json:"from"`
		Text string        `json:"text"`
	}
)

func toPlayerJSON(p room.PlayerView) playerJSON {
	return playerJSON{ID: p.ID, Nickname: p.Nickname, X: p.Position.X, Y: p.Position.Y}
}

// EncodeEvent turns a use case event into a protocol v1 message.
func EncodeEvent(ev usecase.Event) ([]byte, error) {
	typ, data, err := wireData(ev)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{Type: typ, V: ProtocolVersion, Data: raw})
}

func wireData(ev usecase.Event) (string, any, error) {
	if typ, data, ok := cupsWireData(ev); ok {
		return typ, data, nil
	}
	if typ, data, ok := rummyWireData(ev); ok {
		return typ, data, nil
	}
	switch e := ev.(type) {
	case usecase.Joined:
		players := make([]playerJSON, len(e.Players))
		for i, p := range e.Players {
			players[i] = toPlayerJSON(p)
		}
		return TypeRoomJoined, joinedJSON{You: e.You.ID, Map: e.MapName, Players: players, Cups: toCupsSnapshotJSON(e.Cups), Rummy: toRummySnapshotJSON(e.Rummy)}, nil
	case usecase.JoinRejected:
		return TypeRoomJoinRejected, rejectedJSON{Reason: e.Reason}, nil
	case usecase.PlayerJoined:
		return TypePlayerJoined, toPlayerJSON(e.Player), nil
	case usecase.PlayerLeft:
		return TypePlayerLeft, leftJSON{ID: e.ID}, nil
	case usecase.Snapshot:
		players := make([]positionJSON, len(e.Players))
		for i, p := range e.Players {
			players[i] = positionJSON{ID: p.ID, X: p.Position.X, Y: p.Position.Y}
		}
		return TypeSnapshot, snapshotJSON{Tick: e.Tick, Players: players}, nil
	case usecase.ChatMessage:
		return TypeChatMessage, chatJSON{From: e.From, Text: e.Text}, nil
	default:
		return "", nil, fmt.Errorf("no wire format for event %T", ev)
	}
}
