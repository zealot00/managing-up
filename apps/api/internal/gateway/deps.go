package gateway

import (
	"context"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type Principal struct {
	APIKeyID string
	UserID   string
	Username string
	Role     string
}

type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, apiKey string) (*Principal, error)
}

type UsageRecord struct {
	APIKeyID         string
	UserID           string
	Username         string
	Provider         llm.Provider
	Model            llm.Model
	Endpoint         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type UsageRecorder interface {
	RecordUsage(ctx context.Context, record UsageRecord) error
}

type ProviderKeyResolver interface {
	KeyFor(provider llm.Provider) string
}

type Option func(*Server)

func WithAPIKeyValidator(v APIKeyValidator) Option {
	return func(s *Server) {
		s.apiKeyValidator = v
	}
}

func WithUsageRecorder(r UsageRecorder) Option {
	return func(s *Server) {
		s.usageRecorder = r
	}
}

func WithProviderKeyResolver(r ProviderKeyResolver) Option {
	return func(s *Server) {
		s.providerKeyResolver = r
	}
}
