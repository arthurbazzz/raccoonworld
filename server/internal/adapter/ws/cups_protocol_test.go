package ws_test

import (
	"reflect"
	"testing"
	"time"

	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/cups"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// The same JSON is fixed in web/src/net/protocol.test.ts, so client and
// server agree on every three cups message (plan, section 3.4).

func TestDecodeCupsMessages(t *testing.T) {
	tests := []struct {
		raw  string
		want ws.ClientMessage
	}{
		{`{"type":"cups.join","v":1,"data":{}}`, ws.CupsJoinRequest{}},
		{`{"type":"cups.pick","v":1,"data":{"round":2,"slot":1}}`, ws.CupsPickRequest{Round: 2, Slot: 1}},
		{`{"type":"cups.leave","v":1,"data":{}}`, ws.CupsLeaveRequest{}},
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

	in := ws.CupsPickRequest{Round: 2, Slot: 1}.ToInput("c1")
	if in != (usecase.PickInput{PlayerID: "c1", Round: 2, Slot: 1}) {
		t.Fatalf("CupsPickRequest.ToInput() = %+v", in)
	}
}

func TestEncodeCupsEvents(t *testing.T) {
	round := cups.RoundStarted{
		Round: 2, Total: 5, Nut: 1,
		Swaps: []cups.Swap{{0, 1}, {1, 2}},
		Show:  1500 * time.Millisecond, SwapTime: 500 * time.Millisecond, Pick: 10 * time.Second,
	}
	tests := []struct {
		name string
		ev   usecase.Event
		want string
	}{
		{
			"status in the lobby",
			usecase.CupsStatus{Status: cups.StatusChanged{
				Phase: cups.Lobby, Players: []room.PlayerID{"c1"}, Queued: []room.PlayerID{}, LobbyLeft: 6 * time.Second,
			}},
			`{"type":"cups.status","v":1,"data":{"phase":"lobby","players":["c1"],"queued":[],"lobbyMs":6000}}`,
		},
		{
			"status while playing, with a queue",
			usecase.CupsStatus{Status: cups.StatusChanged{
				Phase: cups.Choose, Players: []room.PlayerID{"c1"}, Queued: []room.PlayerID{"c2"},
			}},
			`{"type":"cups.status","v":1,"data":{"phase":"playing","players":["c1"],"queued":["c2"],"lobbyMs":0}}`,
		},
		{
			"status with the scoreboard up",
			usecase.CupsStatus{Status: cups.StatusChanged{Phase: cups.Results}},
			`{"type":"cups.status","v":1,"data":{"phase":"results","players":[],"queued":[],"lobbyMs":0}}`,
		},
		{
			"rejected",
			usecase.CupsRejected{Reason: usecase.CupsTooFar},
			`{"type":"cups.rejected","v":1,"data":{"reason":"too_far"}}`,
		},
		{
			"round",
			usecase.CupsRound{Round: round},
			`{"type":"cups.round","v":1,"data":{"round":2,"total":5,"nut":1,"showMs":1500,
				"swaps":[[0,1],[1,2]],"swapMs":500,"pickMs":10000,"elapsedMs":0}}`,
		},
		{
			"picked, without the cup",
			usecase.CupsPicked{ID: "c1"},
			`{"type":"cups.picked","v":1,"data":{"id":"c1"}}`,
		},
		{
			"reveal",
			usecase.CupsRevealed{Reveal: cups.Revealed{Round: 2, Nut: 0, Picks: []cups.PickResult{
				{ID: "c1", Slot: 0, Picked: true, Hit: true},
				{ID: "c2"},
			}}},
			`{"type":"cups.reveal","v":1,"data":{"round":2,"nut":0,"picks":[
				{"id":"c1","slot":0,"hit":true},{"id":"c2","slot":null,"hit":false}]}}`,
		},
		{
			"results",
			usecase.CupsResults{Results: cups.Finished{Scores: []cups.Score{
				{ID: "c1", Hits: 4, Place: 1}, {ID: "c2", Hits: 4, Place: 1}, {ID: "c3", Hits: 1, Place: 3},
			}}},
			`{"type":"cups.results","v":1,"data":{"scores":[
				{"id":"c1","hits":4,"place":1},{"id":"c2","hits":4,"place":1},{"id":"c3","hits":1,"place":3}]}}`,
		},
		{
			"joined in the middle of a round",
			usecase.Joined{RoomID: "praca", MapName: "praca", You: room.PlayerView{ID: "c3"}, Cups: usecase.CupsSnapshot{
				Status: cups.StatusChanged{Phase: cups.Shuffle, Players: []room.PlayerID{"c1"}},
				Round:  &usecase.CupsRound{Round: round, Elapsed: 2 * time.Second},
			}},
			`{"type":"room.joined","v":1,"data":{"you":"c3","map":"praca","players":[],
				"cups":{"phase":"playing","players":["c1"],"queued":[],"lobbyMs":0,
					"round":{"round":2,"total":5,"nut":1,"showMs":1500,"swaps":[[0,1],[1,2]],
						"swapMs":500,"pickMs":10000,"elapsedMs":2000}},
				"rummy":{"playing":[],"tables":0,"rooms":[]}}}`,
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
