package llm

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported LLM provider")
	ErrInvalidAPIKey       = errors.New("invalid API key")
	ErrInvalidModel        = errors.New("invalid model")
	ErrEmptyResponse       = errors.New("empty response from LLM")
	ErrContextCancelled    = errors.New("context cancelled")
	ErrRateLimited         = errors.New("rate limited")
	ErrAuthentication      = errors.New("authentication failed")
	ErrServerError         = errors.New("server error")
)
