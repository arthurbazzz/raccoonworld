package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// Defaults for Options.
const (
	defaultPingEvery   = 5 * time.Second  // plan §3.3
	defaultPongTimeout = 10 * time.Second // RN5
	writeTimeout       = 5 * time.Second
	sendBuffer         = 64      // messages queued per connection
	readLimit          = 8 << 10 // bytes; the largest client message is a full rummy.move
)

// UseCases are the room operations a connection can trigger.
type UseCases struct {
	Join      *usecase.JoinRoom
	SetInput  *usecase.SetInput
	SendChat  *usecase.SendChat
	Leave     *usecase.LeaveRoom
	KeepAlive *usecase.KeepAlive
	JoinCups  *usecase.JoinCups
	PickCup   *usecase.PickCup
	LeaveCups *usecase.LeaveCups

	OpenTable       *usecase.OpenTable
	CreateTableRoom *usecase.CreateTableRoom
	JoinTableRoom   *usecase.JoinTableRoom
	LeaveTableRoom  *usecase.LeaveTableRoom
	PlayAgain       *usecase.PlayAgain
	PlayTurn        *usecase.PlayTurn
}

// Options configure the handler. Zero values use the defaults.
type Options struct {
	// AllowedOrigins are host patterns (e.g. "localhost:3000") allowed to
	// open a connection from a browser. Same-origin is always allowed.
	AllowedOrigins []string
	PingEvery      time.Duration
	PongTimeout    time.Duration
	Logger         *slog.Logger
}

// Handler upgrades HTTP requests to WebSocket connections, one player each.
type Handler struct {
	hub  *Hub
	uc   UseCases
	opts Options
}

// NewHandler builds the WebSocket endpoint.
func NewHandler(hub *Hub, uc UseCases, opts Options) *Handler {
	if opts.PingEvery == 0 {
		opts.PingEvery = defaultPingEvery
	}
	if opts.PongTimeout == 0 {
		opts.PongTimeout = defaultPongTimeout
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{hub: hub, uc: uc, opts: opts}
}

// ServeHTTP runs one connection until it closes, then removes the player.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.opts.AllowedOrigins})
	if err != nil {
		h.opts.Logger.Info("websocket rejected", "origin", r.Header.Get("Origin"), "error", err)
		return // Accept already wrote the HTTP error
	}
	conn.SetReadLimit(readLimit)

	id := newPlayerID()
	log := h.opts.Logger.With("player", id)
	ctx, cancel := context.WithCancel(r.Context())
	send := make(chan []byte, sendBuffer)

	h.hub.Register(id, send, cancel)
	defer func() {
		cancel()
		h.hub.Unregister(id)
		if err := h.uc.Leave.Execute(id); err != nil {
			log.Warn("leave room", "error", err)
		}
		_ = conn.CloseNow()
		log.Info("connection closed")
	}()
	log.Info("connection opened")

	go h.writeLoop(ctx, cancel, conn, send)
	go h.pingLoop(ctx, cancel, conn, id)
	h.readLoop(ctx, conn, id, log)
}

func (h *Handler) readLoop(ctx context.Context, conn *websocket.Conn, id room.PlayerID, log *slog.Logger) {
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if typ != websocket.MessageText {
			continue
		}
		msg, err := DecodeClient(data)
		if err != nil {
			log.Debug("ignoring invalid message", "error", err)
			continue
		}
		h.dispatch(id, msg, log)
	}
}

func (h *Handler) dispatch(id room.PlayerID, msg ClientMessage, log *slog.Logger) {
	var err error
	switch m := msg.(type) {
	case JoinRequest:
		err = h.uc.Join.Execute(m.ToInput(id))
	case InputRequest:
		err = h.uc.SetInput.Execute(m.ToInput(id))
	case ChatRequest:
		err = h.uc.SendChat.Execute(m.ToInput(id))
	case CupsJoinRequest:
		err = h.uc.JoinCups.Execute(id)
	case CupsPickRequest:
		err = h.uc.PickCup.Execute(m.ToInput(id))
	case CupsLeaveRequest:
		err = h.uc.LeaveCups.Execute(id)
	case RummyOpenRequest:
		err = h.uc.OpenTable.Execute(id)
	case RummyCreateRequest:
		err = h.uc.CreateTableRoom.Execute(id, m.Mode)
	case RummyJoinRequest:
		err = h.uc.JoinTableRoom.Execute(id, m.Room)
	case RummyMoveRequest:
		err = h.uc.PlayTurn.Execute(m.ToInput(id))
	case RummyTurnRequest:
		err = h.uc.PlayTurn.Execute(m.ToInput(id))
	case RummyLeaveRequest:
		err = h.uc.LeaveTableRoom.Execute(id)
	case RummyAgainRequest:
		err = h.uc.PlayAgain.Execute(id)
	case RummySayRequest:
		err = h.uc.SendChat.Execute(m.ToInput(id))
	}
	if err != nil {
		// Rejections were already sent to the player or are silent by spec.
		log.Debug("request refused", "error", err)
	}
	h.keepAlive(id)
}

// keepAlive refreshes presence; before joining there is nothing to refresh.
func (h *Handler) keepAlive(id room.PlayerID) {
	if err := h.uc.KeepAlive.Execute(id); err != nil && !errors.Is(err, room.ErrPlayerNotFound) {
		h.opts.Logger.Warn("keep alive", "player", id, "error", err)
	}
}

func (h *Handler) writeLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, send <-chan []byte) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-send:
			wctx, wcancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Write(wctx, websocket.MessageText, msg)
			wcancel()
			if err != nil {
				return
			}
		}
	}
}

// pingLoop checks the connection is alive. A missing pong closes it, which
// removes the player right away instead of waiting for the room timeout.
func (h *Handler) pingLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, id room.PlayerID) {
	defer cancel()
	ticker := time.NewTicker(h.opts.PingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pctx, pcancel := context.WithTimeout(ctx, h.opts.PongTimeout)
			err := conn.Ping(pctx)
			pcancel()
			if err != nil {
				return
			}
			h.keepAlive(id)
		}
	}
}

func newPlayerID() room.PlayerID {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return room.PlayerID("p-" + hex.EncodeToString(b))
}
