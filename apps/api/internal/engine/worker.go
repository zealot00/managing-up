package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

type Worker struct {
	engine   *ExecutionEngine
	repo     WorkerRepository
	interval time.Duration
	logger   *slog.Logger
}

type WorkerRepository interface {
	ListPendingExecutions() []server.Execution
	ListWaitingApprovalExecutions() []server.Execution
}

func NewWorker(engine *ExecutionEngine, repo WorkerRepository, interval time.Duration) *Worker {
	return &Worker{
		engine:   engine,
		repo:     repo,
		interval: interval,
		logger:   slog.Default(),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("starting execution worker", "interval", w.interval)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("execution worker stopped")
			return
		case <-ticker.C:
			w.processPending()
			w.processWaitingApproval()
		}
	}
}

func (w *Worker) processPending() {
	executions := w.repo.ListPendingExecutions()
	for _, exec := range executions {
		w.logger.Info("processing pending execution", "execution_id", exec.ID)
		if err := w.engine.Run(context.Background(), exec.ID); err != nil {
			w.logger.Error("failed to run execution", "execution_id", exec.ID, "error", err)
		}
	}
}

func (w *Worker) processWaitingApproval() {
	executions := w.repo.ListWaitingApprovalExecutions()
	for _, exec := range executions {
		w.logger.Info("checking waiting execution for resume", "execution_id", exec.ID)
	}
}
