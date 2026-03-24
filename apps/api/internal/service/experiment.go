package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ExperimentRepository defines persistence for experiments.
type ExperimentRepository interface {
	CreateExperiment(exp Experiment) (Experiment, error)
	GetExperiment(id string) (Experiment, bool)
	ListExperiments() []Experiment
	UpdateExperimentStatus(id string, status string) error
}

// ExperimentRunRepository defines persistence for experiment runs.
type ExperimentRunRepository interface {
	CreateExperimentRun(run ExperimentRun) (ExperimentRun, error)
	GetExperimentRun(id string) (ExperimentRun, bool)
	ListExperimentRuns(experimentID string) []ExperimentRun
	UpdateExperimentRun(run ExperimentRun) error
}

// Experiment represents an experiment for comparing agent/skill performance.
type Experiment struct {
	ID          string
	Name        string
	Description string
	TaskIDs     []string
	AgentIDs    []string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ExperimentRun represents an individual run within an experiment.
type ExperimentRun struct {
	ID           string
	ExperimentID string
	TaskID       string
	AgentID      string
	MetricScores map[string]any
	OverallScore float64
	DurationMs   int64
	Status       string
	CreatedAt    time.Time
}

// CreateExperimentRequest is the payload for creating an experiment.
type CreateExperimentRequest struct {
	Name        string
	Description string
	TaskIDs     []string
	AgentIDs    []string
}

// TaskRunner defines the interface for running tasks and evaluating results.
type TaskRunner interface {
	RunTask(ctx context.Context, taskID, agentID string, input map[string]any) (TaskExecution, error)
	EvaluateExecution(ctx context.Context, taskExecID, metricID string) (EvaluationResult, error)
}

// TaskExecution represents a task execution result.
type TaskExecution struct {
	ID         string
	TaskID     string
	AgentID    string
	Status     string
	Input      map[string]any
	Output     map[string]any
	DurationMs int64
	CreatedAt  time.Time
}

// EvaluationResult represents an evaluation result.
type EvaluationResult struct {
	ID              string
	TaskExecutionID string
	MetricID        string
	Score           float64
	Details         map[string]any
	EvaluatedAt     time.Time
}

// ExperimentService orchestrates experiment execution.
type ExperimentService struct {
	experimentRepo ExperimentRepository
	runRepo        ExperimentRunRepository
	taskRepo       TaskRepository
	metricRepo     MetricRepository
	runner         TaskRunner
}

func NewExperimentService(
	experimentRepo ExperimentRepository,
	runRepo ExperimentRunRepository,
	taskRepo TaskRepository,
	metricRepo MetricRepository,
	runner TaskRunner,
) *ExperimentService {
	return &ExperimentService{
		experimentRepo: experimentRepo,
		runRepo:        runRepo,
		taskRepo:       taskRepo,
		metricRepo:     metricRepo,
		runner:         runner,
	}
}

// RunExperiment triggers execution of an experiment.
// It runs all task×agent combinations in parallel using a worker pool.
func (s *ExperimentService) RunExperiment(ctx context.Context, experimentID string) error {
	exp, ok := s.experimentRepo.GetExperiment(experimentID)
	if !ok {
		return fmt.Errorf("experiment not found: %s", experimentID)
	}

	s.experimentRepo.UpdateExperimentStatus(exp.ID, "running")

	// Collect all (task, agent) pairs to run
	type pair struct {
		taskID  string
		agentID string
	}
	var pairs []pair
	for _, taskID := range exp.TaskIDs {
		for _, agentID := range exp.AgentIDs {
			pairs = append(pairs, pair{taskID: taskID, agentID: agentID})
		}
	}

	// Run in parallel (max 10 workers)
	const maxWorkers = 10
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := 0
	_ = completed // suppress unused warning

	for _, p := range pairs {
		wg.Add(1)
		go func(taskID, agentID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s.runSingleTask(ctx, exp.ID, taskID, agentID)

			mu.Lock()
			completed++
			mu.Unlock()
		}(p.taskID, p.agentID)
	}

	wg.Wait()

	s.experimentRepo.UpdateExperimentStatus(exp.ID, "completed")
	return nil
}

// runSingleTask executes one task for one agent and records an ExperimentRun.
func (s *ExperimentService) runSingleTask(ctx context.Context, experimentID, taskID, agentID string) ExperimentRun {
	now := time.Now()
	run := ExperimentRun{
		ID:           fmt.Sprintf("run_%d", time.Now().UnixNano()),
		ExperimentID: experimentID,
		TaskID:       taskID,
		AgentID:      agentID,
		Status:       "running",
		CreatedAt:    now,
	}
	run, _ = s.runRepo.CreateExperimentRun(run)

	// Get task to find test cases
	task, ok := s.taskRepo.GetTask(taskID)
	if !ok {
		run.Status = "failed"
		run.MetricScores = map[string]any{"error": "task not found"}
		s.runRepo.UpdateExperimentRun(run)
		return run
	}

	// Run each test case and collect scores
	metricScores := make(map[string]any)
	var totalScore float64
	var scored int

	for i, tc := range task.TestCases {
		// Execute task
		exec, err := s.runner.RunTask(ctx, taskID, agentID, tc.Input)
		if err != nil {
			continue
		}

		// Evaluate with primary metric
		metricName := task.Scoring.PrimaryMetric
		metrics := s.metricRepo.ListMetrics()
		var metricID string
		for _, m := range metrics {
			if m.Type == metricName {
				metricID = m.ID
				break
			}
		}

		if metricID != "" {
			eval, err := s.runner.EvaluateExecution(ctx, exec.ID, metricID)
			if err == nil {
				metricScores[fmt.Sprintf("case_%d", i)] = eval.Score
				totalScore += eval.Score
				scored++
			}
		}
	}

	if scored > 0 {
		run.OverallScore = totalScore / float64(scored)
	}
	run.MetricScores = metricScores
	run.Status = "completed"
	run.DurationMs = time.Since(now).Milliseconds()
	s.runRepo.UpdateExperimentRun(run)

	return run
}

// ListExperimentResults returns all runs for an experiment with summary stats.
func (s *ExperimentService) ListExperimentResults(experimentID string) ([]ExperimentRun, map[string]any, error) {
	runs := s.runRepo.ListExperimentRuns(experimentID)
	if len(runs) == 0 {
		return runs, nil, nil
	}

	// Compute aggregate stats
	var totalScore float64
	var completed int
	for _, r := range runs {
		if r.Status == "completed" {
			totalScore += r.OverallScore
			completed++
		}
	}

	summary := map[string]any{
		"total_runs": len(runs),
		"completed":  completed,
		"avg_score":  0.0,
	}
	if completed > 0 {
		summary["avg_score"] = totalScore / float64(completed)
	}

	return runs, summary, nil
}

// CreateExperiment creates a new experiment.
func (s *ExperimentService) CreateExperiment(req CreateExperimentRequest) (Experiment, error) {
	if req.Name == "" {
		return Experiment{}, ErrExperimentNameRequired
	}
	exp := Experiment{
		ID:          fmt.Sprintf("exp_%d", time.Now().UnixNano()),
		Name:        req.Name,
		Description: req.Description,
		TaskIDs:     req.TaskIDs,
		AgentIDs:    req.AgentIDs,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	return s.experimentRepo.CreateExperiment(exp)
}

// GetExperiment retrieves an experiment by ID.
func (s *ExperimentService) GetExperiment(id string) (Experiment, bool) {
	return s.experimentRepo.GetExperiment(id)
}

// ListExperiments returns all experiments.
func (s *ExperimentService) ListExperiments() []Experiment {
	return s.experimentRepo.ListExperiments()
}
