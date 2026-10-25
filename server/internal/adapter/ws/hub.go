package ws

import (
	"log/slog"
	"sync"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

var _ usecase.Broadcaster = (*Hub)(nil)

type client struct {
	send   chan<- []byte
	drop   func() // closes the connection
	inRoom bool   // set by Joined, cleared by PlayerLeft
}

// Hub implements usecase.Broadcaster over WebSocket connections. It never
// blocks the room loop: a connection whose send buffer is full is dropped.
type Hub struct {
	logger *slog.Logger

	mu      sync.Mutex
	clients map[room.PlayerID]*client
}

// NewHub creates an empty hub. A nil logger discards logs.
func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Hub{logger: logger, clients: map[room.PlayerID]*client{}}
}

// Register adds a connection. send receives encoded messages; drop is
// called (at most once per overflow) when send is full.
func (h *Hub) Register(id room.PlayerID, send chan<- []byte, drop func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[id] = &client{send: send, drop: drop}
}

// Unregister removes a connection; it receives nothing afterwards.
func (h *Hub) Unregister(id room.PlayerID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
}

// ToPlayer sends ev to one connection. A Joined event marks it as in the
// room, so it starts receiving room broadcasts.
func (h *Hub) ToPlayer(id room.PlayerID, ev usecase.Event) {
	msg, ok := h.encode(ev)
	if !ok {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.clients[id]
	if !ok {
		return
	}
	if _, joined := ev.(usecase.Joined); joined {
		c.inRoom = true
	}
	h.deliver(id, c, msg)
}

// ToRoom sends ev to every connection in the room. A PlayerLeft event
// removes that player from the room first, so they do not receive it.
// This server hosts a single room, so roomID is not needed for routing yet.
func (h *Hub) ToRoom(_ string, ev usecase.Event) {
	msg, ok := h.encode(ev)
	if !ok {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if left, isLeft := ev.(usecase.PlayerLeft); isLeft {
		if c, found := h.clients[left.ID]; found {
			c.inRoom = false
		}
	}
	for id, c := range h.clients {
		if c.inRoom {
			h.deliver(id, c, msg)
		}
	}
}

// deliver must be called with h.mu held.
func (h *Hub) deliver(id room.PlayerID, c *client, msg []byte) {
	select {
	case c.send <- msg:
	default:
		h.logger.Warn("dropping slow connection", "player", id)
		go c.drop()
	}
}

func (h *Hub) encode(ev usecase.Event) ([]byte, bool) {
	msg, err := EncodeEvent(ev)
	if err != nil {
		h.logger.Error("encode event", "event", ev, "error", err)
		return nil, false
	}
	return msg, true
}
