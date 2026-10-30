package ws_test

import (
	"reflect"
	"testing"
	"time"

	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/domain/rummy"
	"raccoons-world/server/internal/usecase"
)

// The same JSON is fixed in the web client's protocol tests, so client and
// server agree on every rummy message (plan, section 3.5).

func TestDecodeRummyMessages(t *testing.T) {
	tests := []struct {
		raw  string
		want ws.ClientMessage
	}{
		{`{"type":"rummy.open","v":1,"data":{}}`, ws.RummyOpenRequest{}},
		{`{"type":"rummy.create","v":1,"data":{"mode":"classic"}}`, ws.RummyCreateRequest{Mode: "classic"}},
		{`{"type":"rummy.join","v":1,"data":{"room":"s1"}}`, ws.RummyJoinRequest{Room: "s1"}},
		{
			`{"type":"rummy.move","v":1,"data":{"turn":7,"melds":[[3,17,40]],"hand":[5,9]}}`,
			ws.RummyMoveRequest{Turn: 7, Melds: [][]int{{3, 17, 40}}, Hand: []int{5, 9}},
		},
		{`{"type":"rummy.end","v":1,"data":{"turn":7}}`, ws.RummyTurnRequest{Action: usecase.ActionEnd, Turn: 7}},
		{`{"type":"rummy.draw","v":1,"data":{"turn":7}}`, ws.RummyTurnRequest{Action: usecase.ActionDraw, Turn: 7}},
		{`{"type":"rummy.undo","v":1,"data":{"turn":7}}`, ws.RummyTurnRequest{Action: usecase.ActionUndo, Turn: 7}},
		{`{"type":"rummy.leave","v":1,"data":{}}`, ws.RummyLeaveRequest{}},
		{`{"type":"rummy.again","v":1,"data":{}}`, ws.RummyAgainRequest{}},
	}
	for _, tt := range tests {
		got, err := ws.DecodeClient([]byte(tt.raw))
		if err != nil {
			t.Fatalf("DecodeClient(%s): %v", tt.raw, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("DecodeClient(%s) = %#v, want %#v", tt.raw, got, tt.want)
		}
	}
}

func TestRummyMoveInput(t *testing.T) {
	in := ws.RummyMoveRequest{Turn: 7, Melds: [][]int{{3, 17, 40}}, Hand: []int{5, -1, 200}}.ToInput("c1")
	want := usecase.TurnInput{
		PlayerID: "c1", Action: usecase.ActionMove, Turn: 7,
		Melds: [][]rummy.TileID{{3, 17, 40}},
		// Out-of-range tiles become an ID no layout holds, so the move is ignored.
		Hand: []rummy.TileID{5, rummy.DeckSize, rummy.DeckSize},
	}
	if !reflect.DeepEqual(in, want) {
		t.Fatalf("ToInput() = %+v, want %+v", in, want)
	}
}

func TestEncodeRummyEvents(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		ev   usecase.Event
		want string
	}{
		{
			"rooms",
			usecase.RummyRooms{Now: now, Rooms: []rummy.RoomInfo{
				{ID: "s1", Mode: rummy.Free, Players: []rummy.Player{{ID: "c1"}, {ID: "c2"}}, Phase: rummy.Lobby, LobbyEnds: now.Add(12 * time.Second)},
				{ID: "s2", Mode: rummy.Classic, Players: []rummy.Player{{ID: "c3"}}, Phase: rummy.Playing},
			}},
			`{"type":"rummy.rooms","v":1,"data":{"rooms":[
				{"id":"s1","mode":"free","players":["c1","c2"],"phase":"lobby","lobbyMs":12000},
				{"id":"s2","mode":"classic","players":["c3"],"phase":"playing","lobbyMs":0}]}}`,
		},
		{
			"status",
			usecase.RummyStatus{Playing: []room.PlayerID{"c1"}, Matches: 1},
			`{"type":"rummy.status","v":1,"data":{"playing":["c1"],"tables":1}}`,
		},
		{
			"empty status",
			usecase.RummyStatus{},
			`{"type":"rummy.status","v":1,"data":{"playing":[],"tables":0}}`,
		},
		{
			"rejected",
			usecase.RummyRejected{Reason: "full"},
			`{"type":"rummy.rejected","v":1,"data":{"reason":"full"}}`,
		},
		{
			"state",
			usecase.RummyState{Room: "s1", Now: now, View: rummy.View{
				Turn: 7, Current: "c1", TurnEnds: now.Add(31 * time.Second), Mode: rummy.Classic,
				Table: [][]rummy.TileID{{3, 4, 5}}, Hand: []rummy.TileID{9, 104},
				Counts: map[room.PlayerID]int{"c1": 2, "c2": 14}, Pool: 75,
				Opened: []room.PlayerID{"c1"},
			}},
			`{"type":"rummy.state","v":1,"data":{"room":"s1","mode":"classic","turn":7,"current":"c1",
				"turnMs":31000,"melds":[[3,4,5]],"hand":[9,104],"counts":{"c1":2,"c2":14},"pool":75,
				"opened":["c1"],"absent":[]}}`,
		},
		{
			"draft",
			usecase.RummyDraft{Turn: 7, Melds: [][]rummy.TileID{{3, 4, 5}, {}}, HandCount: 11},
			`{"type":"rummy.draft","v":1,"data":{"turn":7,"melds":[[3,4,5],[]],"handCount":11}}`,
		},
		{
			"refused turn",
			usecase.RummyTurn{Turn: rummy.TurnEnded{Turn: 7, By: "c1", Result: rummy.ResultInvalid, Reason: "gap", Drew: true}},
			`{"type":"rummy.turn","v":1,"data":{"turn":7,"result":"invalid","reason":"gap","by":"c1","drew":true}}`,
		},
		{
			"results",
			usecase.RummyResults{Results: rummy.MatchOver{Winner: "c1", Scores: []rummy.Score{
				{Player: rummy.Player{ID: "c1"}, Points: 0, Place: 1},
				{Player: rummy.Player{ID: "c2"}, Points: 23, Place: 2},
			}}},
			`{"type":"rummy.results","v":1,"data":{"winner":"c1","scores":[
				{"id":"c1","points":0,"place":1},{"id":"c2","points":23,"place":2}]}}`,
		},
		{
			"results of a tie",
			usecase.RummyResults{Results: rummy.MatchOver{Scores: []rummy.Score{}}},
			`{"type":"rummy.results","v":1,"data":{"winner":null,"scores":[]}}`,
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
