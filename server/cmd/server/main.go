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

	"raccoons-world/server/internal/app"
	"raccoons-world/server/internal/infra/clock"
	"raccoons-world/server/internal/infra/config"
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

	a, err := app.New(app.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		Clock:          clock.System{},
		Logger:         logger,
		CupsSpeedup:    cfg.CupsSpeedup,
	})
	if err != nil {
		return err
	}

	roomDone := make(chan struct{})
	go func() {
		defer close(roomDone)
		a.RunRoom(ctx)
	}()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           a.Handler,
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
