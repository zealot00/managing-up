package gateway

import "testing"

func TestGetModelPricing_O1Lookup(t *testing.T) {
	tests := []struct {
		name           string
		model          string
		wantInputCost  float64
		wantOutputCost float64
	}{
		{
			name:           "exact match gpt-4o",
			model:          "gpt-4o",
			wantInputCost:  0.000005,
			wantOutputCost: 0.000015,
		},
		{
			name:           "lowercase match gpt-4o",
			model:          "gpt-4o",
			wantInputCost:  0.000005,
			wantOutputCost: 0.000015,
		},
		{
			name:           "lowercase gemini",
			model:          "gemini-2.0-flash",
			wantInputCost:  0.000000075,
			wantOutputCost: 0.0000003,
		},
		{
			name:           "mixed case GPT-4O",
			model:          "GPT-4O",
			wantInputCost:  0.000005,
			wantOutputCost: 0.000015,
		},
		{
			name:           "claude haiku lowercase",
			model:          "claude-haiku-3-20250722",
			wantInputCost:  0.00000025,
			wantOutputCost: 0.00000125,
		},
		{
			name:           "unknown model returns default",
			model:          "unknown-model-xyz",
			wantInputCost:  0.000001,
			wantOutputCost: 0.000002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing := GetModelPricing(tt.model)
			if pricing.InputCostPerToken != tt.wantInputCost {
				t.Errorf("InputCostPerToken = %v, want %v", pricing.InputCostPerToken, tt.wantInputCost)
			}
			if pricing.OutputCostPerToken != tt.wantOutputCost {
				t.Errorf("OutputCostPerToken = %v, want %v", pricing.OutputCostPerToken, tt.wantOutputCost)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		wantCost     float64
	}{
		{
			name:         "gpt-4o standard",
			model:        "gpt-4o",
			inputTokens:  1000,
			outputTokens: 500,
			wantCost:     1000*0.000005 + 500*0.000015,
		},
		{
			name:         "case insensitive",
			model:        "GPT-4O",
			inputTokens:  1000,
			outputTokens: 500,
			wantCost:     1000*0.000005 + 500*0.000015,
		},
		{
			name:         "unknown model uses default",
			model:        "unknown",
			inputTokens:  1000,
			outputTokens: 500,
			wantCost:     1000*0.000001 + 500*0.000002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateCost(tt.model, tt.inputTokens, tt.outputTokens)
			if cost != tt.wantCost {
				t.Errorf("CalculateCost = %v, want %v", cost, tt.wantCost)
			}
		})
	}
}
