package evaluation

import (
	"context"
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/service"
)

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

type EvaluationResult struct {
	ID              string
	TaskExecutionID string
	MetricID        string
	Score           float64
	Details         map[string]any
	EvaluatedAt     time.Time
}

type TaskExecutionRepository interface {
	CreateTaskExecution(ex TaskExecution) (TaskExecution, error)
	GetTaskExecution(id string) (TaskExecution, bool)
	UpdateTaskExecution(ex TaskExecution) error
}

type EvaluationRepository interface {
	CreateEvaluation(eval EvaluationResult) error
	GetEvaluation(id string) (EvaluationResult, bool)
	ListEvaluations(taskExecutionID string) []EvaluationResult
}

type EvaluationRunner struct {
	taskRepo   service.TaskRepository
	metricRepo service.MetricRepository
	execRepo   TaskExecutionRepository
	evalRepo   EvaluationRepository
	registry   *EvaluatorRegistry
}

func NewEvaluationRunner(
	taskRepo service.TaskRepository,
	metricRepo service.MetricRepository,
	execRepo TaskExecutionRepository,
	evalRepo EvaluationRepository,
) *EvaluationRunner {
	registry := NewEvaluatorRegistry()
	registry.Register(&ExactMatchEvaluator{})
	registry.Register(NewSemanticSimilarityEvaluator(0.8))

	return &EvaluationRunner{
		taskRepo:   taskRepo,
		metricRepo: metricRepo,
		execRepo:   execRepo,
		evalRepo:   evalRepo,
		registry:   registry,
	}
}

func (r *EvaluationRunner) RegisterJudgeModel(judgeFn PromptBasedJudge) {
	r.registry.Register(NewJudgeModelEvaluator(judgeFn))
}

func (r *EvaluationRunner) RunTask(ctx context.Context, taskID, agentID string, input map[string]any) (TaskExecution, error) {
	task, ok := r.taskRepo.GetTask(taskID)
	if !ok {
		return TaskExecution{}, fmt.Errorf("task not found: %s", taskID)
	}

	exec := TaskExecution{
		ID:        fmt.Sprintf("texec_%d", time.Now().UnixNano()),
		TaskID:    taskID,
		AgentID:   agentID,
		Status:    "running",
		Input:     input,
		CreatedAt: time.Now(),
	}

	exec, err := r.execRepo.CreateTaskExecution(exec)
	if err != nil {
		return TaskExecution{}, fmt.Errorf("failed to create task execution: %w", err)
	}

	output := r.simulateAgentOutput(task, input)
	exec.Output = output
	exec.Status = "completed"

	duration := time.Since(exec.CreatedAt).Milliseconds()
	exec.DurationMs = duration

	r.execRepo.UpdateTaskExecution(exec)

	return exec, nil
}

func (r *EvaluationRunner) EvaluateExecution(ctx context.Context, taskExecID, metricID string) (EvaluationResult, error) {
	exec, ok := r.execRepo.GetTaskExecution(taskExecID)
	if !ok {
		return EvaluationResult{}, fmt.Errorf("task execution not found: %s", taskExecID)
	}

	task, ok := r.taskRepo.GetTask(exec.TaskID)
	if !ok {
		return EvaluationResult{}, fmt.Errorf("task not found: %s", exec.TaskID)
	}

	metric, ok := r.metricRepo.GetMetric(metricID)
	if !ok {
		return EvaluationResult{}, fmt.Errorf("metric not found: %s", metricID)
	}

	evaluator, ok := r.registry.Get(metric.Type)
	if !ok {
		return EvaluationResult{}, fmt.Errorf("evaluator not found for type: %s", metric.Type)
	}

	expected := r.getExpectedForInput(task, exec.Input)
	score, err := evaluator.Evaluate(ctx, exec.Input, expected, exec.Output)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("evaluation failed: %w", err)
	}

	eval := EvaluationResult{
		ID:              fmt.Sprintf("eval_%d", time.Now().UnixNano()),
		TaskExecutionID: taskExecID,
		MetricID:        metricID,
		Score:           score.Value,
		Details:         score.Details,
		EvaluatedAt:     time.Now(),
	}

	if err := r.evalRepo.CreateEvaluation(eval); err != nil {
		return EvaluationResult{}, fmt.Errorf("failed to save evaluation: %w", err)
	}

	return eval, nil
}

func (r *EvaluationRunner) simulateAgentOutput(task service.Task, input map[string]any) map[string]any {
	output := make(map[string]any)
	for _, tc := range task.TestCases {
		for k, v := range tc.Input {
			if input[k] == v {
				output[k] = tc.Expected
			}
		}
	}
	if len(output) == 0 {
		output["result"] = "simulated output"
	}
	return output
}

func (r *EvaluationRunner) getExpectedForInput(task service.Task, input map[string]any) any {
	for _, tc := range task.TestCases {
		match := true
		for k, v := range tc.Input {
			if input[k] != v {
				match = false
				break
			}
		}
		if match {
			return tc.Expected
		}
	}
	return nil
}
