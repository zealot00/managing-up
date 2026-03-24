package llm

import "strings"

// Provider 是 LLM 提供商类型
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderAzure     Provider = "azure"
	ProviderOllama    Provider = "ollama"
	ProviderMinimax   Provider = "minimax"
	ProviderZhipuAI   Provider = "zhipuai"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderBaidu     Provider = "baidu"
	ProviderAlibaba   Provider = "alibaba"
)

// Model 是模型标识符
type Model string

const (
	// OpenAI
	ModelGPT4o     Model = "gpt-4o"
	ModelGPT4oMini Model = "gpt-4o-mini"
	ModelGPT4Turbo Model = "gpt-4-turbo"

	// Anthropic
	ModelClaudeSonnet4 Model = "claude-sonnet-4-20250514"
	ModelClaudeOpus4   Model = "claude-opus-4-20250514"
	ModelClaudeHaiku3  Model = "claude-haiku-3-20250722"

	// Google
	ModelGeminiPro   Model = "gemini-2.0-flash"
	ModelGeminiFlash Model = "gemini-1.5-flash"

	// Ollama (local)
	ModelLlama3  Model = "llama3"
	ModelMistral Model = "mistral"
	ModelQwen25  Model = "qwen2.5"

	// Minimax (国产)
	ModelMinimaxAbot6   Model = "abab6.5s-chat"
	ModelMinimaxM2Ultra Model = "MiniMax-Text-01"
	ModelMinimaxM2Mini  Model = "MiniMax-Text-01-Mini"

	// Zhipu AI / GLM (国产)
	ModelGLM4       Model = "glm-4"
	ModelGLM4Flash  Model = "glm-4-flash"
	ModelGLM4Plus   Model = "glm-4-plus"
	ModelGLM4Vision Model = "glm-4v"
	ModelGLM3Flash  Model = "glm-3-flash"

	// DeepSeek (国产)
	ModelDeepSeekV3    Model = "deepseek-chat"
	ModelDeepSeekCoder Model = "deepseek-coder"

	// Baidu Qianfan / ERNIE (国产)
	ModelERNIE4   Model = "ernie-4.0-8k-latest"
	ModelERNIE35  Model = "ernie-3.5-8k"
	ModelERNIE35V Model = "ernie-3.5-8k-view"
	ModelERNIE4V  Model = "ernie-4.0-8k"

	// Alibaba Tongyi / Qwen (国产)
	ModelQwenMax     Model = "qwen-max"
	ModelQwenPlus    Model = "qwen-plus"
	ModelQwenTurbo   Model = "qwen-turbo"
	ModelQwenMaxLong Model = "qwen-max-long"
)

// Message 是对话消息
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// Response 是 LLM 响应
type Response struct {
	Content      string   `json:"content"`
	Model        Model    `json:"model"`
	Provider     Provider `json:"provider"`
	Usage        Usage    `json:"usage"`
	FinishReason string   `json:"finish_reason"`
}

// Usage 是 token 使用量
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ParseModelString parses a model string and returns the provider and model.
// The model string can be in format "provider:model" (e.g., "openai:gpt-4o")
// or just the model name (e.g., "gpt-4o") in which case the provider is inferred.
func ParseModelString(modelStr string) (Provider, Model, error) {
	// Check if format is "provider:model"
	if strings.Contains(modelStr, ":") {
		parts := strings.SplitN(modelStr, ":", 2)
		provider := Provider(parts[0])
		model := Model(parts[1])
		return provider, model, nil
	}

	// Infer provider from model name prefix
	model := Model(modelStr)
	provider := inferProvider(modelStr)
	return provider, model, nil
}

// inferProvider guesses the provider based on model name prefixes.
func inferProvider(modelStr string) Provider {
	switch {
	case strings.HasPrefix(modelStr, "gpt-4") || strings.HasPrefix(modelStr, "gpt-3.5"):
		return ProviderOpenAI
	case strings.HasPrefix(modelStr, "claude-"):
		return ProviderAnthropic
	case strings.HasPrefix(modelStr, "gemini-"):
		return ProviderGoogle
	case strings.HasPrefix(modelStr, "azure-"):
		return ProviderAzure
	case strings.HasPrefix(modelStr, "llama") || strings.HasPrefix(modelStr, "mistral") || strings.HasPrefix(modelStr, "qwen"):
		return ProviderOllama
	case strings.HasPrefix(modelStr, "abab") || strings.HasPrefix(modelStr, "MiniMax"):
		return ProviderMinimax
	case strings.HasPrefix(modelStr, "glm-"):
		return ProviderZhipuAI
	case strings.HasPrefix(modelStr, "deepseek-"):
		return ProviderDeepSeek
	case strings.HasPrefix(modelStr, "ernie-"):
		return ProviderBaidu
	case strings.HasPrefix(modelStr, "qwen-"):
		return ProviderAlibaba
	default:
		return ProviderOpenAI // default to OpenAI
	}
}
