package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Langa2007/POS/backend/internal/api"
	"github.com/Langa2007/POS/backend/internal/platform/auth"
	"github.com/Langa2007/POS/backend/internal/platform/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	store, err := db.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		logger.Error("apply migrations", "error", err)
		os.Exit(1)
	}
	secret := []byte(os.Getenv("JWT_SECRET"))
	tokens, err := auth.NewTokenManager(secret, 15*time.Minute)
	if err != nil {
		logger.Error("configure access tokens", "error", err)
		os.Exit(1)
	}
	addr := strings.TrimSpace(os.Getenv("POS_ADDR"))
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{Addr: addr, Handler: api.New(store, tokens, logger).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() { logger.Info("api server starting", "addr", addr); errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown server", "error", err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server stopped", "error", err)
			os.Exit(1)
		}
	}
}
