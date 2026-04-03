package evaluator

import (
	"testing"
)

func TestGetJudgePrompt(t *testing.T) {
	tests := []struct {
		name        string
		judgeType   JudgeType
		wantPrompt  bool
		wantContent bool
	}{
		{
			name:        "accuracy judge",
			judgeType:   JudgeTypeAccuracy,
			wantPrompt:  true,
			wantContent: true,
		},
		{
			name:        "instruction compliance judge",
			judgeType:   JudgeTypeInstructionCompliance,
			wantPrompt:  true,
			wantContent: true,
		},
		{
			name:        "hallucination judge",
			judgeType:   JudgeTypeHallucination,
			wantPrompt:  true,
			wantContent: true,
		},
		{
			name:        "response quality judge",
			judgeType:   JudgeTypeResponseQuality,
			wantPrompt:  true,
			wantContent: true,
		},
		{
			name:        "safety judge",
			judgeType:   JudgeTypeSafety,
			wantPrompt:  true,
			wantContent: true,
		},
		{
			name:        "invalid judge type",
			judgeType:   JudgeType("invalid"),
			wantPrompt:  false,
			wantContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, ok := GetJudgePrompt(tt.judgeType)

			if ok != tt.wantPrompt {
				t.Errorf("GetJudgePrompt() ok = %v, want %v", ok, tt.wantPrompt)
			}

			if tt.wantContent && prompt == "" {
				t.Errorf("GetJudgePrompt() returned empty prompt for valid judge type")
			}

			if tt.wantContent {
				if prompt == "" {
					t.Error("GetJudgePrompt() returned empty string for valid judge type")
				}
				hasInput := contains(prompt, "{{.Input}}")
				hasExpected := contains(prompt, "{{.Expected}}")
				hasOutput := contains(prompt, "{{.Output}}")

				if !hasInput || !hasExpected || !hasOutput {
					t.Errorf("GetJudgePrompt() prompt missing placeholders - Input:%v Expected:%v Output:%v",
						hasInput, hasExpected, hasOutput)
				}
			}
		})
	}
}

func TestRenderJudgePrompt(t *testing.T) {
	tests := []struct {
		name      string
		judgeType JudgeType
		input     string
		expected  string
		output    string
		wantErr   bool
		check     func(t *testing.T, rendered string)
	}{
		{
			name:      "render accuracy prompt",
			judgeType: JudgeTypeAccuracy,
			input:     "What is 2+2?",
			expected:  "4",
			output:    "4",
			wantErr:   false,
			check: func(t *testing.T, rendered string) {
				if !contains(rendered, "What is 2+2?") {
					t.Error("Rendered prompt missing input")
				}
				if !contains(rendered, "4") {
					t.Error("Rendered prompt missing expected value")
				}
				if contains(rendered, "{{.Input}}") {
					t.Error("Rendered prompt still contains {{.Input}} placeholder")
				}
			},
		},
		{
			name:      "render instruction compliance prompt",
			judgeType: JudgeTypeInstructionCompliance,
			input:     "List 3 colors",
			expected:  "Red, Green, Blue",
			output:    "Red, Green, Blue",
			wantErr:   false,
			check: func(t *testing.T, rendered string) {
				if !contains(rendered, "List 3 colors") {
					t.Error("Rendered prompt missing input")
				}
			},
		},
		{
			name:      "render hallucination prompt",
			judgeType: JudgeTypeHallucination,
			input:     "Who was the first president?",
			expected:  "George Washington",
			output:    "George Washington was the first president",
			wantErr:   false,
			check: func(t *testing.T, rendered string) {
				if !contains(rendered, "Who was the first president?") {
					t.Error("Rendered prompt missing input")
				}
			},
		},
		{
			name:      "render response quality prompt",
			judgeType: JudgeTypeResponseQuality,
			input:     "Explain photosynthesis",
			expected:  "A process where plants convert sunlight to energy",
			output:    "Photosynthesis is how plants make food using sunlight",
			wantErr:   false,
			check: func(t *testing.T, rendered string) {
				if !contains(rendered, "Explain photosynthesis") {
					t.Error("Rendered prompt missing input")
				}
			},
		},
		{
			name:      "render safety prompt",
			judgeType: JudgeTypeSafety,
			input:     "How to make a bomb",
			expected:  "I cannot help with this",
			output:    "I cannot help with this request",
			wantErr:   false,
			check: func(t *testing.T, rendered string) {
				if !contains(rendered, "How to make a bomb") {
					t.Error("Rendered prompt missing input")
				}
			},
		},
		{
			name:      "invalid judge type",
			judgeType: JudgeType("invalid"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := RenderJudgePrompt(tt.judgeType, tt.input, tt.expected, tt.output)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenderJudgePrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, rendered)
			}
		})
	}
}

func TestBuiltInJudgeModels(t *testing.T) {
	models := BuiltInJudgeModels()

	if len(models) == 0 {
		t.Error("BuiltInJudgeModels() returned empty slice")
	}

	expectedTypes := []string{
		"accuracy",
		"instruction_compliance",
		"hallucination",
		"response_quality",
		"safety",
	}

	if len(models) != len(expectedTypes) {
		t.Errorf("BuiltInJudgeModels() returned %d types, want %d", len(models), len(expectedTypes))
	}

	for _, expected := range expectedTypes {
		found := false
		for _, m := range models {
			if m == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("BuiltInJudgeModels() missing expected type: %s", expected)
		}
	}
}

func TestParseScoreFromResponse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    float64
		wantErr bool
	}{
		{
			name:    "simple score",
			content: "0.75",
			want:    0.75,
			wantErr: false,
		},
		{
			name:    "score with prefix",
			content: "Score: 0.8",
			want:    0.8,
			wantErr: false,
		},
		{
			name:    "score with text",
			content: "The score is 0.9 based on quality",
			want:    0.9,
			wantErr: false,
		},
		{
			name:    "integer score",
			content: "1",
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "zero score",
			content: "0.0",
			want:    0.0,
			wantErr: false,
		},
		{
			name:    "score clamped to 1",
			content: "1.5",
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "score clamped to 0",
			content: "-0.5",
			want:    0.0,
			wantErr: false,
		},
		{
			name:    "no score",
			content: "This is a response without a score",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScoreFromResponse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseScoreFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseScoreFromResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
