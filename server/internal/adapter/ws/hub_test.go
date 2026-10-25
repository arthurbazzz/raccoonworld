package ws_test

import (
	"strings"
	"testing"
	"time"

	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

type testClient struct {
	send chan []byte
	slow chan struct{}
}

func register(h *ws.Hub, id room.PlayerID, buffer int) *testClient {
	c := &testClient{send: make(chan []byte, buffer), slow: make(chan struct{}, 1)}
	h.Register(id, c.send, func() { c.slow <- struct{}{} })
	return c
}

func (c *testClient) next(t *testing.T) string {
	t.Helper()
	select {
	case m := <-c.send:
		return string(m)
	default:
		t.Fatal("no message queued")
		return ""
	}
}

func (c *testClient) empty(t *testing.T) {
	t.Helper()
	select {
	case m := <-c.send:
		t.Fatalf("unexpected message %s", m)
	default:
	}
}

func joined(id room.PlayerID) usecase.Joined {
	return usecase.Joined{RoomID: "praca", MapName: "praca", You: room.PlayerView{ID: id}}
}

func TestToPlayerReachesOnlyThatConnection(t *testing.T) {
	h := ws.NewHub(nil)
	a := register(h, "a", 4)
	b := register(h, "b", 4)

	h.ToPlayer("a", usecase.JoinRejected{Reason: usecase.RejectRoomFull})

	if msg := a.next(t); !strings.Contains(msg, `"room.join_rejected"`) {
		t.Fatalf("a got %s", msg)
	}
	b.empty(t)
}

func TestToRoomReachesOnlyJoinedConnections(t *testing.T) {
	h := ws.NewHub(nil)
	a := register(h, "a", 4)
	b := register(h, "b", 4) // connected but never joined

	h.ToPlayer("a", joined("a"))
	_ = a.next(t)

	h.ToRoom("praca", usecase.ChatMessage{From: "a", Text: "oi"})

	if msg := a.next(t); !strings.Contains(msg, `"chat.message"`) {
		t.Fatalf("a got %s", msg)
	}
	b.empty(t)
}

func TestPlayerLeftStopsRoomBroadcasts(t *testing.T) {
	h := ws.NewHub(nil)
	a := register(h, "a", 4)
	b := register(h, "b", 4)
	h.ToPlayer("a", joined("a"))
	h.ToPlayer("b", joined("b"))
	_, _ = a.next(t), b.next(t)

	h.ToRoom("praca", usecase.PlayerLeft{ID: "a"}) // a timed out

	b.next(t)
	a.empty(t) // a does not hear about itself leaving
	h.ToRoom("praca", usecase.ChatMessage{From: "b", Text: "oi"})
	b.next(t)
	a.empty(t)
}

func TestSlowConnectionIsDropped(t *testing.T) {
	h := ws.NewHub(nil)
	a := register(h, "a", 1)
	h.ToPlayer("a", joined("a")) // fills the buffer

	h.ToRoom("praca", usecase.ChatMessage{From: "a", Text: "oi"})

	select {
	case <-a.slow:
	case <-time.After(time.Second):
		t.Fatal("slow connection was not dropped")
	}
}

func TestUnregisteredConnectionGetsNothing(t *testing.T) {
	h := ws.NewHub(nil)
	a := register(h, "a", 4)
	h.ToPlayer("a", joined("a"))
	_ = a.next(t)
	h.Unregister("a")

	h.ToRoom("praca", usecase.ChatMessage{From: "x", Text: "oi"})
	h.ToPlayer("a", usecase.PlayerLeft{ID: "x"})
	a.empty(t)
}
