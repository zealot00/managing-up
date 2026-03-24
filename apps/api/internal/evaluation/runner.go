package evaluation

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/llm"
	"github.com/zealot/managing-up/apps/api/internal/runtime"
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
	taskRepo     service.TaskRepository
	metricRepo   service.MetricRepository
	execRepo     TaskExecutionRepository
	evalRepo     EvaluationRepository
	registry     *EvaluatorRegistry
	traceEmitter runtime.TraceEmitter
}

func NewEvaluationRunner(
	taskRepo service.TaskRepository,
	metricRepo service.MetricRepository,
	execRepo TaskExecutionRepository,
	evalRepo EvaluationRepository,
	traceEmitter runtime.TraceEmitter,
) *EvaluationRunner {
	registry := NewEvaluatorRegistry()
	registry.Register(&ExactMatchEvaluator{})
	registry.Register(NewSemanticSimilarityEvaluator(0.8))

	return &EvaluationRunner{
		taskRepo:     taskRepo,
		metricRepo:   metricRepo,
		execRepo:     execRepo,
		evalRepo:     evalRepo,
		registry:     registry,
		traceEmitter: traceEmitter,
	}
}

func (r *EvaluationRunner) RegisterJudgeModel(judgeFn PromptBasedJudge) {
	r.registry.Register(NewJudgeModelEvaluator(judgeFn))
}

func (r *EvaluationRunner) emitEvent(ctx context.Context, eventType runtime.EventType, data any) {
	if r.traceEmitter == nil {
		return
	}
	event := runtime.TraceEvent{
		ID:          runtime.GenerateTraceID(),
		ExecutionID: "", // will be set by caller
		EventType:   eventType,
		EventData:   runtime.MustBuildEventData(data),
		Timestamp:   time.Now(),
	}
	r.traceEmitter.Emit(ctx, event)
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

	r.emitEvent(ctx, runtime.EventExecutionStarted, runtime.ExecutionStartedData{
		SkillID:   taskID,
		SkillName: agentID,
		Input:     input,
	})

	startTime := time.Now()
	output, err := callLLM(ctx, task, input)
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		exec.Status = "failed"
		exec.Output = map[string]any{"error": err.Error()}
		r.execRepo.UpdateTaskExecution(exec)
		r.emitEvent(ctx, runtime.EventExecutionFailed, map[string]any{
			"error": err.Error(),
		})
		return exec, fmt.Errorf("LLM call failed: %w", err)
	}

	var inputTokens int
	if tokens, ok := output["tokens"].(int); ok {
		inputTokens = tokens
	}
	r.emitEvent(ctx, runtime.EventLLMCall, runtime.LLMCallData{
		Model:       task.Execution.Model,
		Output:      fmt.Sprintf("%v", output["result"]),
		InputTokens: inputTokens,
		DurationMs:  durationMs,
	})

	exec.Output = output
	exec.Status = "completed"

	duration := time.Since(exec.CreatedAt).Milliseconds()
	exec.DurationMs = duration

	r.execRepo.UpdateTaskExecution(exec)

	r.emitEvent(ctx, runtime.EventExecutionSucceeded, map[string]any{
		"output":      output,
		"duration_ms": duration,
	})

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

	expected := ""
	for _, tc := range task.TestCases {
		match := true
		for k, v := range tc.Input {
			if fmt.Sprintf("%v", exec.Input[k]) != fmt.Sprintf("%v", v) {
				match = false
				break
			}
		}
		if match {
			expected = fmt.Sprintf("%v", tc.Expected)
			break
		}
	}
	if expected == "" {
		expected = fmt.Sprintf("%v", task.Gold.Data)
	}
	score, err := evaluator.Evaluate(ctx, exec.Input, expected, exec.Output)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("evaluation failed: %w", err)
	}

	r.emitEvent(ctx, runtime.EventLLMCall, runtime.LLMCallData{
		Model:  task.Execution.Model,
		Output: fmt.Sprintf("score: %.4f", score.Value),
	})

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

// getLLMClient returns an LLM client configured for the task.
func getLLMClient(task service.Task) (llm.Client, error) {
	provider, model, err := llm.ParseModelString(task.Execution.Model)
	if err != nil {
		return nil, fmt.Errorf("invalid model string %q: %w", task.Execution.Model, err)
	}

	apiKey := os.Getenv("LLM_API_KEY")
	// For Ollama, no API key needed
	if provider == llm.ProviderOllama {
		apiKey = ""
	}

	client, err := llm.NewClient(provider, model, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}
	return client, nil
}

// callLLM invokes the LLM and returns the output text.
func callLLM(ctx context.Context, task service.Task, input map[string]any) (map[string]any, error) {
	client, err := getLLMClient(task)
	if err != nil {
		return nil, err
	}

	// Build the prompt from input
	prompt := buildPrompt(task, input)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	opts := []llm.Option{
		llm.WithTemperature(float32(task.Execution.Temperature)),
		llm.WithMaxTokens(task.Execution.MaxTokens),
	}

	resp, err := client.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	output := map[string]any{
		"result": resp.Content,
		"model":  task.Execution.Model,
		"tokens": resp.Usage.TotalTokens,
	}

	return output, nil
}

// buildPrompt creates a prompt from task input and gold/expected data.
func buildPrompt(task service.Task, input map[string]any) string {
	// Format input as a readable string
	var parts []string
	for k, v := range input {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	inputStr := joinLines(parts)

	prompt := fmt.Sprintf("Task: %s\nInput:\n%s\n", task.Description, inputStr)

	// If we have gold/expected data, include it as reference
	if len(task.TestCases) > 0 {
		for i, tc := range task.TestCases {
			if fmt.Sprintf("%v", tc.Input) == fmt.Sprintf("%v", input) {
				prompt += fmt.Sprintf("\nExpected output format: %v", tc.Expected)
				break
			}
			if i == 0 {
				prompt += fmt.Sprintf("\nExample expected: %v", tc.Expected)
			}
		}
	}

	prompt += "\nProvide your output."
	return prompt
}

func joinLines(parts []string) string {
	if len(parts) == 0 {
		return "(no input)"
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += "\n" + p
	}
	return result
}
