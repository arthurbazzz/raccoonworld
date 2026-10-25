package ws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
	"raccoons-world/server/maps"
)

type wallClock struct{}

func (wallClock) Now() time.Time { return time.Now() }

// newServer runs the WebSocket handler for a real Praça room loop.
func newServer(t *testing.T, origins []string) *httptest.Server {
	t.Helper()
	tiles, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		t.Fatal(err)
	}
	hub := ws.NewHub(nil)
	runner := usecase.NewRoomRunner(room.New("praca", tiles), hub, wallClock{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); runner.Run(ctx, nil) }()

	h := ws.NewHandler(hub, ws.UseCases{
		Join:      usecase.NewJoinRoom(runner, hub, wallClock{}),
		SetInput:  usecase.NewSetInput(runner),
		SendChat:  usecase.NewSendChat(runner, hub, wallClock{}),
		Leave:     usecase.NewLeaveRoom(runner, hub),
		KeepAlive: usecase.NewKeepAlive(runner, wallClock{}),
	}, ws.Options{AllowedOrigins: origins})

	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close(); cancel(); <-done })
	return srv
}

// dial opens a WebSocket with the given Origin and returns the HTTP status
// of the handshake response.
func dial(t *testing.T, srv *httptest.Server, origin string) (*websocket.Conn, int, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	header := http.Header{}
	if origin != "" {
		header.Set("Origin", origin)
	}
	conn, resp, err := websocket.Dial(ctx, "ws"+srv.URL[len("http"):], &websocket.DialOptions{HTTPHeader: header})
	status := 0
	if resp != nil {
		status = resp.StatusCode
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}
	return conn, status, err
}

func TestHandlerRejectsUnknownOrigin(t *testing.T) {
	srv := newServer(t, []string{"localhost:3000"})

	conn, status, err := dial(t, srv, "http://evil.example")
	if err == nil {
		_ = conn.CloseNow()
		t.Fatal("dial from a foreign origin succeeded")
	}
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", status)
	}
}

func TestHandlerAcceptsAllowedOrigin(t *testing.T) {
	srv := newServer(t, []string{"localhost:3000"})

	conn, _, err := dial(t, srv, "http://localhost:3000")
	if err != nil {
		t.Fatalf("dial from allowed origin: %v", err)
	}
	_ = conn.CloseNow()
}
