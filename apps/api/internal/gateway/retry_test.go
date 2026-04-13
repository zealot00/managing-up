package gateway

import (
	"errors"
	"testing"
)

func TestIsNonRetryableError_CaseInsensitiveAndStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "uppercase invalid api key",
			err:  errors.New("Invalid API Key for provider"),
			want: true,
		},
		{
			name: "http 401",
			err:  errors.New("MiniMax API returned status 401: unauthorized"),
			want: true,
		},
		{
			name: "timeout should remain retryable",
			err:  errors.New("context deadline exceeded"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNonRetryableError(tt.err)
			if got != tt.want {
				t.Fatalf("isNonRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}
