package gateway

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNoOpBudgetChecker_CheckBudget(t *testing.T) {
	checker := &NoOpBudgetChecker{}
	ctx := context.Background()

	allowed, remaining, err := checker.CheckBudget(ctx, "user1", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed to be true")
	}
	if remaining != 0 {
		t.Errorf("expected remaining 0, got %d", remaining)
	}
}

func TestNoOpBudgetChecker_DecrementBudget(t *testing.T) {
	checker := &NoOpBudgetChecker{}
	ctx := context.Background()

	remaining, err := checker.DecrementBudget(ctx, "user1", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 0 {
		t.Errorf("expected remaining 0, got %d", remaining)
	}
}

func TestNoOpBudgetChecker_GetBudget(t *testing.T) {
	checker := &NoOpBudgetChecker{}
	ctx := context.Background()

	used, limit, err := checker.GetBudget(ctx, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if used != 0 {
		t.Errorf("expected used 0, got %d", used)
	}
	if limit != 0 {
		t.Errorf("expected limit 0, got %d", limit)
	}
}

func TestNoOpBudgetChecker_ResetBudget(t *testing.T) {
	checker := &NoOpBudgetChecker{}
	ctx := context.Background()

	err := checker.ResetBudget(ctx, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEstimateRequestTokens(t *testing.T) {
	tests := []struct {
		name          string
		contentLength int64
		expectedMin   int
		expectedMax   int
	}{
		{"zero length", 0, 0, 200},
		{"1000 bytes", 1000, 200, 300},
		{"4000 bytes", 4000, 900, 1100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{}
			req.ContentLength = tt.contentLength
			result := estimateRequestTokens(req)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("estimateRequestTokens(%d) = %d, want between %d and %d",
					tt.contentLength, result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestSecondsToEndOfDay(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	result := secondsToEndOfDay(now)
	if result <= 0 || result > 24*60*60 {
		t.Errorf("secondsToEndOfDay() = %d, want > 0 and <= 86400", result)
	}
}

func TestSecondsToEndOfMonth(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	result := secondsToEndOfMonth(now)
	if result <= 0 || result > 31*24*60*60 {
		t.Errorf("secondsToEndOfMonth() = %d, want > 0 and <= 2678400", result)
	}
}
