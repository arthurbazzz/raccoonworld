// Package app wires the server together: room, use cases, WebSocket
// adapter and HTTP routes. cmd/server and the end-to-end smoke tests both
// build the server through New, so they run the same code.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"raccoons-world/server/internal/adapter/httpapi"
	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
	"raccoons-world/server/maps"
)

// Options configure an App.
type Options struct {
	AllowedOrigins []string
	Clock          usecase.Clock
	Logger         *slog.Logger // nil discards logs
}

// App is a ready-to-serve Raccoons World server with a single room.
type App struct {
	Handler http.Handler
	runner  *usecase.RoomRunner
}

// New builds the Praça room and the HTTP handler around it.
func New(opts Options) (*App, error) {
	tiles, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		return nil, err
	}

	hub := ws.NewHub(opts.Logger)
	runner := usecase.NewRoomRunner(room.New("praca", tiles), hub, opts.Clock)
	wsHandler := ws.NewHandler(hub, ws.UseCases{
		Join:      usecase.NewJoinRoom(runner, hub, opts.Clock),
		SetInput:  usecase.NewSetInput(runner),
		SendChat:  usecase.NewSendChat(runner, hub, opts.Clock),
		Leave:     usecase.NewLeaveRoom(runner, hub),
		KeepAlive: usecase.NewKeepAlive(runner, opts.Clock),
	}, ws.Options{AllowedOrigins: opts.AllowedOrigins, Logger: opts.Logger})

	router := httpapi.NewRouter(httpapi.Config{
		WebSocket:      wsHandler,
		Maps:           map[string][]byte{tiles.Name(): maps.Praca},
		AllowedOrigins: opts.AllowedOrigins,
	})
	return &App{Handler: router, runner: runner}, nil
}

// RunRoom runs the room loop at usecase.TickInterval until ctx is done.
func (a *App) RunRoom(ctx context.Context) {
	ticker := time.NewTicker(usecase.TickInterval)
	defer ticker.Stop()
	a.runner.Run(ctx, ticker.C)
}
