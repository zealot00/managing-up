package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/config"
	engine "github.com/zealot/managing-up/apps/api/internal/engine"
	"github.com/zealot/managing-up/apps/api/internal/repository/postgres"
	"github.com/zealot/managing-up/apps/api/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()
	server.SetLogger(slog.Default())
	srv := server.New(cfg)

	if cfg.Database.Enabled() {
		repo, err := postgres.New(cfg.Database.DSN)
		if err != nil {
			log.Fatalf("postgres initialization failed: %v", err)
		}

		srv = server.NewWithRepository(cfg, repo, repo.Close)

		execEngine := engine.NewExecutionEngine(repo, engine.NewToolGateway())
		worker := engine.NewWorker(execEngine, repo, 2*time.Second)
		go worker.Start(context.Background())
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Start()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server exited with error: %v", err)
		}
	case <-stopCtx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}
