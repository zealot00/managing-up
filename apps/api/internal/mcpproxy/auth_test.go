package mcpproxy

import (
	"context"
	"testing"
)

func TestPrincipalFromContext(t *testing.T) {
	t.Parallel()

	t.Run("with principal", func(t *testing.T) {
		p := &Principal{APIKeyID: "key_123", UserID: "user_456"}
		ctx := context.WithValue(context.Background(), principalKey, p)
		got := PrincipalFromContext(ctx)
		if got == nil {
			t.Fatal("expected principal, got nil")
		}
		if got.APIKeyID != "key_123" {
			t.Errorf("expected APIKeyID key_123, got %s", got.APIKeyID)
		}
		if got.UserID != "user_456" {
			t.Errorf("expected UserID user_456, got %s", got.UserID)
		}
	})

	t.Run("without principal", func(t *testing.T) {
		got := PrincipalFromContext(context.Background())
		if got != nil {
			t.Fatal("expected nil, got principal")
		}
	})

	t.Run("with wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), principalKey, "not a principal")
		got := PrincipalFromContext(ctx)
		if got != nil {
			t.Fatal("expected nil for wrong type, got principal")
		}
	})
}

func TestExtractBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"standard bearer", "Bearer skhub_abc123", "skhub_abc123"},
		{"no prefix", "skhub_abc123", "skhub_abc123"},
		{"empty string", "", ""},
		{"bearer only", "Bearer ", ""},
		{"with extra spaces", "Bearer  token123 ", " token123 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBearerToken(tt.header)
			if got != tt.expected {
				t.Errorf("extractBearerToken(%q) = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}
