// Command server runs the Raccoons World backend: one room (the Praça),
// a WebSocket endpoint at /ws and a health check at /health.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"raccoons-world/server/internal/adapter/httpapi"
	"raccoons-world/server/internal/adapter/ws"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/infra/clock"
	"raccoons-world/server/internal/infra/config"
	"raccoons-world/server/internal/usecase"
	"raccoons-world/server/maps"
)

const shutdownTimeout = 5 * time.Second

func main() {
	cfg := config.FromEnv()
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if err := run(cfg, logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tiles, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		return err
	}

	clk := clock.System{}
	hub := ws.NewHub(logger)
	runner := usecase.NewRoomRunner(room.New("praca", tiles), hub, clk)

	ticker := time.NewTicker(usecase.TickInterval)
	defer ticker.Stop()
	roomDone := make(chan struct{})
	go func() {
		defer close(roomDone)
		runner.Run(ctx, ticker.C)
	}()

	wsHandler := ws.NewHandler(hub, ws.UseCases{
		Join:      usecase.NewJoinRoom(runner, hub, clk),
		SetInput:  usecase.NewSetInput(runner),
		SendChat:  usecase.NewSendChat(runner, hub, clk),
		Leave:     usecase.NewLeaveRoom(runner, hub),
		KeepAlive: usecase.NewKeepAlive(runner, clk),
	}, ws.Options{AllowedOrigins: cfg.AllowedOrigins, Logger: logger})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.NewRouter(wsHandler),
		ReadHeaderTimeout: 5 * time.Second,
		// Request contexts derive from ctx, so a shutdown signal also closes
		// hijacked WebSocket connections, which Shutdown does not track.
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.Addr, "allowedOrigins", cfg.AllowedOrigins)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		stop()
		<-roomDone
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = srv.Shutdown(shutdownCtx)
	<-roomDone
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
