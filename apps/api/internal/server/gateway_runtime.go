package server

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/gateway"
	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type gatewayAPIKeyValidator struct {
	repo Repository
}

func (v gatewayAPIKeyValidator) ValidateAPIKey(ctx context.Context, apiKey string) (*gateway.Principal, error) {
	if apiKey == "" {
		return nil, errors.New("empty api key")
	}
	keyHash := HashGatewayAPIKey(apiKey)
	key, ok := v.repo.GetGatewayAPIKeyByHash(keyHash)
	if !ok || key.RevokedAt != nil {
		return nil, errors.New("invalid api key")
	}
	_ = v.repo.TouchGatewayAPIKeyLastUsed(key.ID, time.Now().UTC())
	return &gateway.Principal{
		APIKeyID: key.ID,
		UserID:   key.UserID,
		Username: key.Username,
		Role:     key.Role,
	}, nil
}

type gatewayUsageRecorder struct {
	repo Repository
}

func (r gatewayUsageRecorder) RecordUsage(ctx context.Context, record gateway.UsageRecord) error {
	cost := gateway.CalculateCost(string(record.Model), record.PromptTokens, record.CompletionTokens)

	event := GatewayUsageEvent{
		ID:               generateRequestID(),
		APIKeyID:         record.APIKeyID,
		UserID:           record.UserID,
		Username:         record.Username,
		Provider:         string(record.Provider),
		Model:            string(record.Model),
		Endpoint:         record.Endpoint,
		PromptTokens:     record.PromptTokens,
		CompletionTokens: record.CompletionTokens,
		TotalTokens:      record.TotalTokens,
		Cost:             cost,
		CreatedAt:        time.Now().UTC(),
	}
	return r.repo.CreateGatewayUsageEvent(event)
}

type staticProviderKeyResolver struct {
	providerKeys map[llm.Provider]string
	defaultKey   string
}

func (r staticProviderKeyResolver) KeyFor(userID string, provider llm.Provider) string {
	if key := r.providerKeys[provider]; key != "" {
		return key
	}
	return r.defaultKey
}

type DBProviderKeyResolver struct {
	repo        Repository
	envResolver gateway.ProviderKeyResolver
}

func (r *DBProviderKeyResolver) KeyFor(userID string, provider llm.Provider) string {
	keys := r.repo.ListGatewayProviderKeys(userID)
	for _, k := range keys {
		if k.Provider == string(provider) && k.IsEnabled {
			return k.EncryptedKey
		}
	}
	return r.envResolver.KeyFor(userID, provider)
}

func buildProviderKeyResolverFromEnv() gateway.ProviderKeyResolver {
	providerKeys := map[llm.Provider]string{
		llm.ProviderOpenAI:    os.Getenv("GATEWAY_OPENAI_API_KEY"),
		llm.ProviderAnthropic: os.Getenv("GATEWAY_ANTHROPIC_API_KEY"),
		llm.ProviderGoogle:    os.Getenv("GATEWAY_GOOGLE_API_KEY"),
		llm.ProviderAzure:     os.Getenv("GATEWAY_AZURE_API_KEY"),
		llm.ProviderOllama:    os.Getenv("GATEWAY_OLLAMA_API_KEY"),
		llm.ProviderMinimax:   os.Getenv("GATEWAY_MINIMAX_API_KEY"),
		llm.ProviderZhipuAI:   os.Getenv("GATEWAY_ZHIPUAI_API_KEY"),
		llm.ProviderDeepSeek:  os.Getenv("GATEWAY_DEEPSEEK_API_KEY"),
		llm.ProviderBaidu:     os.Getenv("GATEWAY_BAIDU_API_KEY"),
		llm.ProviderAlibaba:   os.Getenv("GATEWAY_ALIBABA_API_KEY"),
	}

	return staticProviderKeyResolver{
		providerKeys: providerKeys,
		defaultKey:   os.Getenv("GATEWAY_DEFAULT_UPSTREAM_API_KEY"),
	}
}

func buildDBProviderKeyResolver(repo Repository) gateway.ProviderKeyResolver {
	envResolver := buildProviderKeyResolverFromEnv()
	return &DBProviderKeyResolver{
		repo:        repo,
		envResolver: envResolver,
	}
}
