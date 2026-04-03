package engine

import (
	"testing"
)

func TestNoOpTokenizer_Count(t *testing.T) {
	tokenizer := NewNoOpTokenizer()

	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 2},
		{"hello world", 3},
		{"four chars", 3},
	}

	for _, tt := range tests {
		result := tokenizer.Count(tt.input)
		if result != tt.expected {
			t.Errorf("Count(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestNoOpTokenizer_CountMessages(t *testing.T) {
	tokenizer := NewNoOpTokenizer()

	messages := []Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	count := tokenizer.CountMessages(messages)
	if count <= 0 {
		t.Errorf("CountMessages() = %d, want > 0", count)
	}
}

func TestContextTruncator_TruncateIfNeeded(t *testing.T) {
	tokenizer := NewNoOpTokenizer()
	truncator := NewContextTruncator(tokenizer, 100)

	messages := []Message{
		{Role: "system", Content: "You are helpful"},
	}

	result, truncated, err := truncator.TruncateIfNeeded(messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if truncated {
		t.Error("expected not truncated for small input")
	}
	if len(result) != len(messages) {
		t.Errorf("expected %d messages, got %d", len(messages), len(result))
	}
}

func TestContextTruncator_NeedsSummarization(t *testing.T) {
	tokenizer := NewNoOpTokenizer()
	truncator := NewContextTruncator(tokenizer, 100)

	messages := []Message{
		{Role: "system", Content: "You are helpful"},
	}

	needs := truncator.NeedsSummarization(messages)
	if needs {
		t.Error("expected false for small input")
	}
}
