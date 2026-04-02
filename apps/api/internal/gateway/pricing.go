package gateway

import (
	"strings"
)

type ModelPricing struct {
	InputCostPerToken  float64
	OutputCostPerToken float64
}

var modelPricing = map[string]ModelPricing{
	"gpt-4o":                   {InputCostPerToken: 0.000005, OutputCostPerToken: 0.000015},
	"gpt-4o-mini":              {InputCostPerToken: 0.00000015, OutputCostPerToken: 0.0000006},
	"gpt-4-turbo":              {InputCostPerToken: 0.00001, OutputCostPerToken: 0.00003},
	"o1-preview":               {InputCostPerToken: 0.000015, OutputCostPerToken: 0.00006},
	"o1-mini":                  {InputCostPerToken: 0.000003, OutputCostPerToken: 0.000012},
	"claude-sonnet-4-20250514": {InputCostPerToken: 0.000003, OutputCostPerToken: 0.000015},
	"claude-opus-4-20250514":   {InputCostPerToken: 0.000015, OutputCostPerToken: 0.000075},
	"claude-haiku-3-20250722":  {InputCostPerToken: 0.00000025, OutputCostPerToken: 0.00000125},
	"gemini-2.0-flash":         {InputCostPerToken: 0.000000075, OutputCostPerToken: 0.0000003},
	"gemini-1.5-flash":         {InputCostPerToken: 0.000000075, OutputCostPerToken: 0.0000003},
	"deepseek-chat":            {InputCostPerToken: 0.00000014, OutputCostPerToken: 0.00000028},
	"deepseek-coder":           {InputCostPerToken: 0.00000014, OutputCostPerToken: 0.00000028},
	"glm-4":                    {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000002},
	"glm-4-flash":              {InputCostPerToken: 0.0000001, OutputCostPerToken: 0.0000001},
	"glm-4-plus":               {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000002},
	"ernie-4.0-8k-latest":      {InputCostPerToken: 0.0000012, OutputCostPerToken: 0.0000012},
	"ernie-3.5-8k":             {InputCostPerToken: 0.00000028, OutputCostPerToken: 0.00000028},
	"qwen-max":                 {InputCostPerToken: 0.0000016, OutputCostPerToken: 0.0000048},
	"qwen-plus":                {InputCostPerToken: 0.0000004, OutputCostPerToken: 0.0000012},
	"qwen-turbo":               {InputCostPerToken: 0.00000006, OutputCostPerToken: 0.00000006},
	"abab6.5s-chat":            {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000001},
	"MiniMax-Text-01":          {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000001},
}

func GetModelPricing(model string) ModelPricing {
	if pricing, ok := modelPricing[model]; ok {
		return pricing
	}

	lowerModel := strings.ToLower(model)
	for key, pricing := range modelPricing {
		if strings.ToLower(key) == lowerModel {
			return pricing
		}
	}

	return ModelPricing{
		InputCostPerToken:  0.000001,
		OutputCostPerToken: 0.000002,
	}
}

func CalculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing := GetModelPricing(model)
	inputCost := float64(inputTokens) * pricing.InputCostPerToken
	outputCost := float64(outputTokens) * pricing.OutputCostPerToken
	return inputCost + outputCost
}
