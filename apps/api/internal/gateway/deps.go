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
	ClientName       string
	Provider         llm.Provider
	Model            llm.Model
	Endpoint         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
}

type UsageRecorder interface {
	RecordUsage(ctx context.Context, record UsageRecord) error
}

type ProviderKeyResolver interface {
	KeyFor(userID string, provider llm.Provider) string
}

type BudgetChecker interface {
	CheckBudget(ctx context.Context, key string, tokens int) (bool, int, error)
	DecrementBudget(ctx context.Context, key string, tokens int) (int, error)
	GetBudget(ctx context.Context, key string) (used, limit int, err error)
	ResetBudget(ctx context.Context, key string) error
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

func WithRouter(r LLMRouter) Option {
	return func(s *Server) {
		s.router = r
	}
}

func WithBudgetChecker(b BudgetChecker) Option {
	return func(s *Server) {
		s.budgetChecker = b
	}
}
