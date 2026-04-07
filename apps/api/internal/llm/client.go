package llm

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
)

var (
	scannerBufferSize    = getEnvInt("GATEWAY_SCANNER_BUFFER_SIZE", 10*1024*1024)
	scannerMaxBufferSize = getEnvInt("GATEWAY_SCANNER_MAX_BUFFER_SIZE", 50*1024*1024)
)

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// Client 是统一接口
type Client interface {
	Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error)
	GenerateStream(ctx context.Context, messages []Message, opts ...Option) (StreamReader, error)
	Provider() Provider
	Model() Model
}

// StreamReader is the interface for streaming LLM responses
type StreamReader interface {
	// Recv returns the next chunk from the stream.
	// Returns io.EOF when the stream is finished.
	Recv() (*StreamChunk, error)
}

// StreamChunk represents a single chunk in a streaming response
type StreamChunk struct {
	// Content is the text content of this chunk
	Content string `json:"content"`
	// Done indicates if this is the final chunk
	Done bool `json:"done"`
	// FinishReason indicates why the generation finished (if Done is true)
	FinishReason string `json:"finish_reason,omitempty"`
	// Usage contains token usage information (only in the final chunk)
	Usage *Usage `json:"usage,omitempty"`
	// Model contains the model used
	Model Model `json:"model,omitempty"`
}

// Option 是生成选项
type Option func(*GenerateOptions)

type GenerateOptions struct {
	Temperature float32
	MaxTokens   int
	TopP        float32
	StopWords   []string
	JSONMode    bool
}

func WithTemperature(t float32) Option {
	return func(o *GenerateOptions) {
		o.Temperature = t
	}
}

func WithMaxTokens(m int) Option {
	return func(o *GenerateOptions) {
		o.MaxTokens = m
	}
}

func WithTopP(p float32) Option {
	return func(o *GenerateOptions) {
		o.TopP = p
	}
}

func WithStopWords(words ...string) Option {
	return func(o *GenerateOptions) {
		o.StopWords = words
	}
}

func WithJSONResponse() Option {
	return func(o *GenerateOptions) {
		o.JSONMode = true
	}
}

// NewClient 根据 provider 创建 client
func NewClient(provider Provider, model Model, apiKey string) (Client, error) {
	switch provider {
	case ProviderOpenAI:
		return NewOpenAIClient(model, apiKey), nil
	case ProviderAnthropic:
		return NewAnthropicClient(model, apiKey), nil
	case ProviderGoogle:
		return NewGoogleClient(model, apiKey), nil
	case ProviderAzure:
		return NewAzureClient(model, apiKey), nil
	case ProviderOllama:
		return NewOllamaClient(model, apiKey), nil
	case ProviderMinimax:
		return NewMinimaxClient(model, apiKey), nil
	case ProviderZhipuAI:
		return NewZhipuAIClient(model, apiKey), nil
	case ProviderDeepSeek:
		return NewDeepSeekClient(model, apiKey), nil
	case ProviderBaidu:
		return NewBaiduClient(model, apiKey), nil
	case ProviderAlibaba:
		return NewAlibabaClient(model, apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// Config 从环境变量读取配置
type Config struct {
	Provider Provider
	Model    Model
	APIKey   string
	BaseURL  string
}

func ConfigFromEnv() Config {
	return Config{
		Provider: Provider(os.Getenv("LLM_PROVIDER")),
		Model:    Model(os.Getenv("LLM_MODEL")),
		APIKey:   os.Getenv("LLM_API_KEY"),
		BaseURL:  os.Getenv("LLM_BASE_URL"),
	}
}

func newLargeBufferScanner(body io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(body)
	buf := make([]byte, scannerBufferSize)
	scanner.Buffer(buf, scannerMaxBufferSize)
	return scanner
}
