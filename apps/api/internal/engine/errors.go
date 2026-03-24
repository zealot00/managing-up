package engine

import "fmt"

type RuntimeError struct {
	ExecutionID string
	StepID      string
	Err         error
}

func (e *RuntimeError) Error() string {
	if e.ExecutionID != "" && e.StepID != "" {
		return fmt.Sprintf("runtime error at execution %s step %s: %v", e.ExecutionID, e.StepID, e.Err)
	}
	if e.ExecutionID != "" {
		return fmt.Sprintf("runtime error at execution %s: %v", e.ExecutionID, e.Err)
	}
	return fmt.Sprintf("runtime error: %v", e.Err)
}

func (e *RuntimeError) Unwrap() error {
	return e.Err
}

func NewExecutionError(execID, stepID string, err error) *RuntimeError {
	return &RuntimeError{ExecutionID: execID, StepID: stepID, Err: err}
}

func NewStepError(stepID string, err error) *RuntimeError {
	return &RuntimeError{StepID: stepID, Err: err}
}

var (
	ErrExecutionNotFound    = fmt.Errorf("execution not found")
	ErrSkillNotFound        = fmt.Errorf("skill not found")
	ErrSkillVersionNotFound = fmt.Errorf("skill version not found")
	ErrInvalidSpec          = fmt.Errorf("invalid skill spec")
	ErrNoPublishedVersion   = fmt.Errorf("no published version for skill")
	ErrStepNotFound         = fmt.Errorf("step not found in skill spec")
)
