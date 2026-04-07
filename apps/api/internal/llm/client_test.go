package llm

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		model    Model
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "OpenAI client",
			provider: ProviderOpenAI,
			model:    ModelGPT4o,
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "Anthropic client",
			provider: ProviderAnthropic,
			model:    ModelClaudeSonnet4,
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "Google client",
			provider: ProviderGoogle,
			model:    ModelGeminiPro,
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "Azure client",
			provider: ProviderAzure,
			model:    ModelGPT4o,
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "Ollama client",
			provider: ProviderOllama,
			model:    ModelLlama3,
			apiKey:   "",
			wantErr:  false,
		},
		{
			name:     "Unsupported provider",
			provider: Provider("unknown"),
			model:    ModelGPT4o,
			apiKey:   "test-key",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.provider, tt.model, tt.apiKey)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Errorf("NewClient() returned nil client")
				return
			}
			if client.Provider() != tt.provider {
				t.Errorf("Client.Provider() = %v, want %v", client.Provider(), tt.provider)
			}
			if client.Model() != tt.model {
				t.Errorf("Client.Model() = %v, want %v", client.Model(), tt.model)
			}
		})
	}
}

func TestClientGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	client := NewOpenAIClient(ModelGPT4o, "test-key")
	resp, err := client.Generate(context.Background(), []Message{
		{Role: "user", Content: "hello"},
	})
	if err != nil {
		t.Skipf("skipping: OpenAI API not available: %v", err)
	}
	if resp == nil {
		t.Errorf("Generate() expected response, got nil")
	}
}

func TestOllamaClientNoAPIKey(t *testing.T) {
	client, err := NewClient(ProviderOllama, ModelLlama3, "")
	if err != nil {
		t.Errorf("NewClient() for Ollama failed: %v", err)
	}
	if client == nil {
		t.Errorf("NewClient() returned nil for Ollama")
	}
}
