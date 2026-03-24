package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

type Repository interface {
	GetSkillVersionForExecution(skillID string) (server.SkillVersion, bool)
	UpdateExecutionStatus(id string, status string, stepID string, endedAt *time.Time, durationMs *int64) error
	CreateExecutionStep(step server.ExecutionStep) error
	GetExecutionForResume(id string) (server.Execution, bool)
}

type ExecutionEngine struct {
	repo         Repository
	toolGateway  *ToolGateway
	parser       *SkillSpecParser
	logger       *slog.Logger
	traceEmitter TraceEmitter
}

func NewExecutionEngine(repo Repository, toolGateway *ToolGateway) *ExecutionEngine {
	return &ExecutionEngine{
		repo:         repo,
		toolGateway:  toolGateway,
		parser:       NewSkillSpecParser(),
		logger:       slog.Default(),
		traceEmitter: &noOpEmitter{},
	}
}

func NewExecutionEngineWithTracer(repo Repository, toolGateway *ToolGateway, tracer TraceEmitter) *ExecutionEngine {
	return &ExecutionEngine{
		repo:         repo,
		toolGateway:  toolGateway,
		parser:       NewSkillSpecParser(),
		logger:       slog.Default(),
		traceEmitter: tracer,
	}
}

func (e *ExecutionEngine) Run(ctx context.Context, executionID string) error {
	e.logger.Info("starting execution", "execution_id", executionID)

	execution, ok := e.repo.GetExecutionForResume(executionID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrExecutionNotFound, executionID)
	}

	e.repo.UpdateExecutionStatus(executionID, "running", execution.CurrentStepID, nil, nil)

	e.traceEmitter.Emit(ctx, TraceEvent{
		ID:          GenerateTraceID(),
		ExecutionID: executionID,
		EventType:   EventExecutionStarted,
		EventData:   MustBuildEventData(ExecutionStartedData{SkillID: execution.SkillID, SkillName: execution.SkillName, Input: execution.Input, Triggered: execution.TriggeredBy}),
		Timestamp:   time.Now(),
	})

	sv, ok := e.repo.GetSkillVersionForExecution(execution.SkillID)
	if !ok {
		return e.failExecution(executionID, "", fmt.Errorf("%w: skill %s", ErrSkillNotFound, execution.SkillID))
	}

	spec, err := e.parser.Parse(sv.SpecYaml)
	if err != nil {
		return e.failExecution(executionID, "", fmt.Errorf("failed to parse skill spec: %w", err))
	}

	if err := e.parser.ValidateStepOrder(spec); err != nil {
		return e.failExecution(executionID, "", err)
	}

	stepIndex := e.findStepIndex(spec, execution.CurrentStepID)
	if stepIndex == -1 {
		stepIndex = 0
	}

	for i := stepIndex; i < len(spec.Steps); i++ {
		step := spec.Steps[i]

		select {
		case <-ctx.Done():
			return e.failExecution(executionID, step.ID, ctx.Err())
		default:
		}

		switch step.Type {
		case "tool":
			if err := e.executeToolStep(ctx, executionID, step, spec.Inputs, execution.Input); err != nil {
				if spec.OnFailure != nil && spec.OnFailure.Action == OnFailureActionContinue {
					e.logger.Warn("step failed but continuing due to on_failure policy", "step_id", step.ID, "error", err)
					continue
				}
				return err
			}
		case "approval":
			if err := e.handleApprovalStep(ctx, executionID, step); err != nil {
				return err
			}
			return nil
		case "condition":
			e.logger.Info("condition step - evaluating", "step_id", step.ID)
		}
	}

	return e.succeedExecution(executionID)
}

func (e *ExecutionEngine) Resume(ctx context.Context, executionID string) error {
	e.logger.Info("resuming execution", "execution_id", executionID)

	execution, ok := e.repo.GetExecutionForResume(executionID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrExecutionNotFound, executionID)
	}

	if execution.Status != "waiting_approval" && execution.Status != "running" {
		return fmt.Errorf("cannot resume execution with status: %s", execution.Status)
	}

	return e.Run(ctx, executionID)
}

func (e *ExecutionEngine) executeToolStep(ctx context.Context, executionID string, step Step, inputs []Input, execInput map[string]any) error {
	toolRef := step.ToolRef
	if toolRef == "" {
		return fmt.Errorf("%w: %s", ErrStepNotFound, step.ID)
	}

	timeout := step.TimeoutSeconds
	if timeout == 0 {
		timeout = 30
	}

	maxAttempts := 1
	backoff := 0
	if step.RetryPolicy != nil {
		maxAttempts = step.RetryPolicy.MaxAttempts
		backoff = step.RetryPolicy.BackoffSeconds
		if maxAttempts == 0 {
			maxAttempts = 1
		}
	}

	e.logger.Info("executing tool step",
		"execution_id", executionID,
		"step_id", step.ID,
		"tool_ref", toolRef,
		"timeout_seconds", timeout,
		"max_attempts", maxAttempts)

	startedAt := time.Now()
	stepRecordID := fmt.Sprintf("step_%s_%d", executionID, time.Now().UnixNano())
	e.repo.CreateExecutionStep(server.ExecutionStep{
		ID:          stepRecordID,
		ExecutionID: executionID,
		StepID:      step.ID,
		Status:      StepStatusRunning,
		ToolRef:     toolRef,
		StartedAt:   startedAt,
		AttemptNo:   1,
	})

	e.traceEmitter.Emit(ctx, TraceEvent{
		ID:          GenerateTraceID(),
		ExecutionID: executionID,
		StepID:      step.ID,
		EventType:   EventStepStarted,
		EventData:   MustBuildEventData(StepEventData{StepID: step.ID, StepType: step.Type, ToolRef: toolRef, AttemptNo: 1}),
		Timestamp:   startedAt,
	})

	toolInput := make(map[string]any)
	for k, v := range step.With {
		toolInput[k] = e.interpolateInput(v, inputs, execInput)
	}

	inv := GatewayToolInvocation{
		ExecutionID:    executionID,
		StepID:         step.ID,
		ToolRef:        toolRef,
		Input:          toolInput,
		TimeoutSeconds: timeout,
	}

	var result *GatewayToolResult
	var err error

	if maxAttempts > 1 {
		result, err = e.toolGateway.InvokeWithRetry(ctx, inv, maxAttempts, backoff)
	} else {
		result, err = e.toolGateway.Invoke(ctx, inv)
	}

	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()

	if err != nil {
		e.repo.CreateExecutionStep(server.ExecutionStep{
			ID:          stepRecordID,
			ExecutionID: executionID,
			StepID:      step.ID,
			Status:      StepStatusFailed,
			ToolRef:     toolRef,
			StartedAt:   startedAt,
			EndedAt:     &endedAt,
			DurationMs:  durationMs,
			Error:       err.Error(),
			AttemptNo:   1,
		})
		e.traceEmitter.Emit(ctx, TraceEvent{
			ID:          GenerateTraceID(),
			ExecutionID: executionID,
			StepID:      step.ID,
			EventType:   EventStepFailed,
			EventData:   MustBuildEventData(ToolEventData{StepID: step.ID, ToolRef: toolRef, Error: err.Error(), DurationMs: durationMs}),
			Timestamp:   endedAt,
		})
		return e.failExecution(executionID, step.ID, fmt.Errorf("tool invocation failed: %w", err))
	}

	if result.Status != "succeeded" {
		e.repo.CreateExecutionStep(server.ExecutionStep{
			ID:          stepRecordID,
			ExecutionID: executionID,
			StepID:      step.ID,
			Status:      StepStatusFailed,
			ToolRef:     toolRef,
			StartedAt:   startedAt,
			EndedAt:     &endedAt,
			DurationMs:  durationMs,
			Error:       result.Error,
			AttemptNo:   1,
		})
		e.traceEmitter.Emit(ctx, TraceEvent{
			ID:          GenerateTraceID(),
			ExecutionID: executionID,
			StepID:      step.ID,
			EventType:   EventStepFailed,
			EventData:   MustBuildEventData(ToolEventData{StepID: step.ID, ToolRef: toolRef, Error: result.Error, DurationMs: durationMs}),
			Timestamp:   endedAt,
		})
		return e.failExecution(executionID, step.ID, fmt.Errorf("tool returned status: %s", result.Status))
	}

	e.repo.CreateExecutionStep(server.ExecutionStep{
		ID:          stepRecordID,
		ExecutionID: executionID,
		StepID:      step.ID,
		Status:      StepStatusSucceeded,
		ToolRef:     toolRef,
		StartedAt:   startedAt,
		EndedAt:     &endedAt,
		DurationMs:  durationMs,
		Output:      result.Output,
		AttemptNo:   1,
	})

	e.traceEmitter.Emit(ctx, TraceEvent{
		ID:          GenerateTraceID(),
		ExecutionID: executionID,
		StepID:      step.ID,
		EventType:   EventStepSucceeded,
		EventData:   MustBuildEventData(ToolEventData{StepID: step.ID, ToolRef: toolRef, Output: result.Output, DurationMs: durationMs}),
		Timestamp:   endedAt,
	})

	e.repo.UpdateExecutionStatus(executionID, "running", step.ID, nil, nil)
	return nil
}

func (e *ExecutionEngine) handleApprovalStep(ctx context.Context, executionID string, step Step) error {
	e.logger.Info("approval step - pausing execution",
		"execution_id", executionID,
		"step_id", step.ID,
		"approver_group", step.ApproverGroup)

	e.repo.UpdateExecutionStatus(executionID, "waiting_approval", step.ID, nil, nil)

	e.traceEmitter.Emit(ctx, TraceEvent{
		ID:          GenerateTraceID(),
		ExecutionID: executionID,
		StepID:      step.ID,
		EventType:   EventApprovalRequested,
		EventData:   MustBuildEventData(ApprovalEventData{StepID: step.ID, ApproverGroup: step.ApproverGroup, Message: step.Message}),
		Timestamp:   time.Now(),
	})
	return nil
}

func (e *ExecutionEngine) succeedExecution(executionID string) error {
	endedAt := time.Now()
	e.repo.UpdateExecutionStatus(executionID, "succeeded", "completed", &endedAt, nil)
	e.logger.Info("execution succeeded", "execution_id", executionID)
	e.traceEmitter.Emit(context.Background(), TraceEvent{
		ID:          GenerateTraceID(),
		ExecutionID: executionID,
		EventType:   EventExecutionSucceeded,
		EventData:   MustBuildEventData(StateChangeData{From: "running", To: "succeeded"}),
		Timestamp:   endedAt,
	})
	return nil
}

func (e *ExecutionEngine) failExecution(executionID, stepID string, err error) error {
	e.logger.Error("execution failed",
		"execution_id", executionID,
		"step_id", stepID,
		"error", err)
	endedAt := time.Now()
	e.repo.UpdateExecutionStatus(executionID, "failed", stepID, &endedAt, nil)
	e.traceEmitter.Emit(context.Background(), TraceEvent{
		ID:          GenerateTraceID(),
		ExecutionID: executionID,
		StepID:      stepID,
		EventType:   EventExecutionFailed,
		EventData:   MustBuildEventData(StateChangeData{From: "running", To: "failed", Step: stepID}),
		Timestamp:   endedAt,
	})
	return NewExecutionError(executionID, stepID, err)
}

func (e *ExecutionEngine) findStepIndex(spec *SkillSpec, currentStepID string) int {
	if currentStepID == "" || currentStepID == "queued" || currentStepID == "resumed_after_approval" {
		return 0
	}
	for i, step := range spec.Steps {
		if step.ID == currentStepID {
			return i
		}
	}
	return 0
}

func (e *ExecutionEngine) interpolateInput(value string, inputs []Input, execInput map[string]any) string {
	for _, input := range inputs {
		placeholder := fmt.Sprintf("{{ inputs.%s }}", input.Name)
		if value == placeholder {
			if v, ok := execInput[input.Name]; ok {
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return value
}
