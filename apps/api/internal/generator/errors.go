package generator

import "errors"

var (
	ErrInvalidSpec      = errors.New("invalid skill spec")
	ErrValidationFailed = errors.New("validation failed")
	ErrLLMCallFailed    = errors.New("llm call failed")
)
