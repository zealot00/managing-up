package engine

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

const maxConcurrentExecutions = 50

type Worker struct {
	engine      *ExecutionEngine
	repo        WorkerRepository
	interval    time.Duration
	logger      *slog.Logger
	sem         chan struct{}
	wg          sync.WaitGroup
	activeExecs sync.Map
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
		sem:      make(chan struct{}, maxConcurrentExecutions),
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
			w.wg.Wait()
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
		execID := exec.ID
		if _, loaded := w.activeExecs.LoadOrStore(execID, true); loaded {
			w.logger.Debug("execution already in progress, skipping", "execution_id", execID)
			continue
		}

		select {
		case w.sem <- struct{}{}:
			w.wg.Add(1)
			go func(exec server.Execution) {
				defer func() {
					<-w.sem
					w.wg.Done()
					w.activeExecs.Delete(exec.ID)
				}()
				w.logger.Info("processing pending execution", "execution_id", exec.ID)
				if err := w.engine.Run(context.Background(), exec.ID); err != nil {
					w.logger.Error("failed to run execution", "execution_id", exec.ID, "error", err)
				}
			}(exec)
		default:
			w.logger.Warn("worker pool full, dropping execution", "execution_id", execID)
		}
	}
}

func (w *Worker) processWaitingApproval() {
	executions := w.repo.ListWaitingApprovalExecutions()
	for _, exec := range executions {
		w.logger.Info("resuming execution after approval", "execution_id", exec.ID)
		if err := w.engine.Resume(context.Background(), exec.ID); err != nil {
			w.logger.Error("failed to resume execution", "execution_id", exec.ID, "error", err)
		}
	}
}
