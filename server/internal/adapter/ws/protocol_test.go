package ws_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

func TestDecodeClientMessages(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want ws.ClientMessage
	}{
		{
			"join",
			`{"type":"room.join","v":1,"data":{"nickname":"rocky"}}`,
			ws.JoinRequest{Nickname: "rocky"},
		},
		{
			"input",
			`{"type":"player.input","v":1,"data":{"up":true,"down":false,"left":false,"right":true,"seq":7}}`,
			ws.InputRequest{Up: true, Right: true, Seq: 7},
		},
		{
			"chat",
			`{"type":"chat.send","v":1,"data":{"text":"olá!"}}`,
			ws.ChatRequest{Text: "olá!"},
		},
		{
			"unknown fields are ignored",
			`{"type":"chat.send","v":1,"data":{"text":"oi","color":"red"},"extra":1}`,
			ws.ChatRequest{Text: "oi"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ws.DecodeClient([]byte(tt.raw))
			if err != nil {
				t.Fatalf("DecodeClient() error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("DecodeClient() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecodeClientErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want error
	}{
		{"not json", `{`, ws.ErrMalformed},
		{"data not an object", `{"type":"chat.send","v":1,"data":"oi"}`, ws.ErrMalformed},
		{"missing data", `{"type":"chat.send","v":1}`, ws.ErrMalformed},
		{"unknown type", `{"type":"player.teleport","v":1,"data":{}}`, ws.ErrUnknownType},
		{"server-only type", `{"type":"room.snapshot","v":1,"data":{}}`, ws.ErrUnknownType},
		{"missing version", `{"type":"chat.send","data":{"text":"oi"}}`, ws.ErrUnsupportedVersion},
		{"future version", `{"type":"chat.send","v":2,"data":{"text":"oi"}}`, ws.ErrUnsupportedVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ws.DecodeClient([]byte(tt.raw)); !errors.Is(err, tt.want) {
				t.Fatalf("DecodeClient() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestRequestsMapToUseCaseInputs(t *testing.T) {
	const id room.PlayerID = "c1"

	if got := (ws.JoinRequest{Nickname: "rocky"}).ToInput(id); got != (usecase.JoinInput{PlayerID: id, Nickname: "rocky"}) {
		t.Fatalf("JoinRequest.ToInput() = %+v", got)
	}
	in := (ws.InputRequest{Up: true, Left: true}).ToInput(id)
	if in != (usecase.SetInputInput{PlayerID: id, Up: true, Left: true}) {
		t.Fatalf("InputRequest.ToInput() = %+v", in)
	}
	if got := (ws.ChatRequest{Text: "oi"}).ToInput(id); got != (usecase.SendChatInput{PlayerID: id, Text: "oi"}) {
		t.Fatalf("ChatRequest.ToInput() = %+v", got)
	}
}

func TestEncodeEvents(t *testing.T) {
	rocky := room.PlayerView{ID: "c1", Nickname: "rocky", Position: room.Vec{X: 480, Y: 400}}
	bandit := room.PlayerView{ID: "c2", Nickname: "bandit", Position: room.Vec{X: 488.5, Y: 400}}

	tests := []struct {
		name string
		ev   usecase.Event
		want string
	}{
		{
			"joined",
			usecase.Joined{RoomID: "praca", MapName: "praca", You: rocky, Players: []room.PlayerView{bandit, rocky}},
			`{"type":"room.joined","v":1,"data":{"you":"c1","map":"praca","players":[
				{"id":"c2","nickname":"bandit","x":488.5,"y":400},
				{"id":"c1","nickname":"rocky","x":480,"y":400}],
				"cups":{"phase":"idle","players":[],"queued":[],"lobbyMs":0,"round":null}}}`,
		},
		{
			"join rejected",
			usecase.JoinRejected{Reason: usecase.RejectRoomFull},
			`{"type":"room.join_rejected","v":1,"data":{"reason":"room_full"}}`,
		},
		{
			"player joined",
			usecase.PlayerJoined{Player: rocky},
			`{"type":"player.joined","v":1,"data":{"id":"c1","nickname":"rocky","x":480,"y":400}}`,
		},
		{
			"player left",
			usecase.PlayerLeft{ID: "c1"},
			`{"type":"player.left","v":1,"data":{"id":"c1"}}`,
		},
		{
			"snapshot",
			usecase.Snapshot{RoomID: "praca", Tick: 42, Players: []room.PlayerView{bandit}},
			`{"type":"room.snapshot","v":1,"data":{"tick":42,"players":[{"id":"c2","x":488.5,"y":400}]}}`,
		},
		{
			"chat message",
			usecase.ChatMessage{From: "c1", Text: "olá!"},
			`{"type":"chat.message","v":1,"data":{"from":"c1","text":"olá!"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ws.EncodeEvent(tt.ev)
			if err != nil {
				t.Fatalf("EncodeEvent() error: %v", err)
			}
			assertJSONEqual(t, got, tt.want)
		})
	}
}

func TestEncodeEmptyListsAsArrays(t *testing.T) {
	got, err := ws.EncodeEvent(usecase.Joined{RoomID: "praca", MapName: "praca", You: room.PlayerView{ID: "c1"}})
	if err != nil {
		t.Fatal(err)
	}
	var msg struct {
		Data struct {
			Players json.RawMessage `json:"players"`
		} `json:"data"`
	}
	if err := json.Unmarshal(got, &msg); err != nil {
		t.Fatal(err)
	}
	if string(msg.Data.Players) != "[]" {
		t.Fatalf("players = %s, want [] (clients iterate it)", msg.Data.Players)
	}
}

func assertJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("bad expectation: %v", err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("JSON mismatch\n got: %s\nwant: %s", got, want)
	}
}
