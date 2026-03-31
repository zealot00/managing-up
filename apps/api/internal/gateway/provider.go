package gateway

import "strings"

type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderZhipuAI   Provider = "zhipuai"
	ProviderOllama    Provider = "ollama"
)

func ParseModelString(modelStr string) (Provider, string, error) {
	if strings.Contains(modelStr, ":") {
		parts := strings.SplitN(modelStr, ":", 2)
		return Provider(parts[0]), parts[1], nil
	}

	return inferProvider(modelStr), modelStr, nil
}

func inferProvider(modelStr string) Provider {
	switch {
	case strings.HasPrefix(modelStr, "gpt-") || strings.HasPrefix(modelStr, "o1-"):
		return ProviderOpenAI
	case strings.HasPrefix(modelStr, "claude-"):
		return ProviderAnthropic
	case strings.HasPrefix(modelStr, "gemini-"):
		return ProviderGoogle
	case strings.HasPrefix(modelStr, "deepseek-"):
		return ProviderDeepSeek
	case strings.HasPrefix(modelStr, "glm-"):
		return ProviderZhipuAI
	case strings.HasPrefix(modelStr, "llama") || strings.HasPrefix(modelStr, "mistral") || strings.HasPrefix(modelStr, "qwen2"):
		return ProviderOllama
	default:
		return ProviderOpenAI
	}
}
