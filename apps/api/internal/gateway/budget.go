package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type BudgetConfig struct {
	MonthlyLimit   int
	DailyLimit     int
	AlertThreshold float64
}

type RedisBudgetChecker struct {
	client *redis.Client
	config BudgetConfig
	script *redis.Script
}

// Lua script for atomic check and decrement of daily/monthly budget
// KEYS[1] = daily key, KEYS[2] = monthly key
// ARGV[1] = tokens to check/decrement
// ARGV[2] = daily limit
// ARGV[3] = monthly limit
// ARGV[4] = current timestamp (ms)
// ARGV[5] = daily TTL (seconds to end of day)
// ARGV[6] = monthly TTL (seconds to end of month)
// Returns: [allowed (0/1), remaining daily, remaining monthly]
const budgetCheckAndDecrementScript = `
local daily_key = KEYS[1]
local monthly_key = KEYS[2]
local tokens = tonumber(ARGV[1])
local daily_limit = tonumber(ARGV[2])
local monthly_limit = tonumber(ARGV[3])
local now_ms = tonumber(ARGV[4])
local daily_ttl = tonumber(ARGV[5])
local monthly_ttl = tonumber(ARGV[6])

-- Get current usage
local daily_used = tonumber(redis.call('GET', daily_key) or '0')
local monthly_used = tonumber(redis.call('GET', monthly_key) or '0')

-- Check if budget would be exceeded
if daily_used + tokens > daily_limit then
    local remaining_daily = math.max(0, daily_limit - daily_used)
    return {0, remaining_daily, monthly_limit - monthly_used}
end

if monthly_used + tokens > monthly_limit then
    local remaining_monthly = math.max(0, monthly_limit - monthly_used)
    return {0, daily_limit - daily_used, remaining_monthly}
end

-- Atomically increment both counters
local new_daily = redis.call('INCRBY', daily_key, tokens)
local new_monthly = redis.call('INCRBY', monthly_key, tokens)

-- Set TTL if this is the first increment
if new_daily == tokens then
    redis.call('PEXPIRE', daily_key, daily_ttl)
end
if new_monthly == tokens then
    redis.call('PEXPIRE', monthly_key, monthly_ttl)
end

return {1, daily_limit - new_daily, monthly_limit - new_monthly}
`

// NewRedisBudgetChecker creates a new Redis-based budget checker
func NewRedisBudgetChecker(client *redis.Client, config BudgetConfig) *RedisBudgetChecker {
	return &RedisBudgetChecker{
		client: client,
		config: config,
		script: redis.NewScript(budgetCheckAndDecrementScript),
	}
}

// CheckBudget checks if the request is allowed and returns remaining daily budget
func (r *RedisBudgetChecker) CheckBudget(ctx context.Context, key string, tokens int) (bool, int, error) {
	dailyKey := fmt.Sprintf("budget:daily:%s", key)
	monthlyKey := fmt.Sprintf("budget:monthly:%s", key)

	now := time.Now()
	dailyTTL := secondsToEndOfDay(now)
	monthlyTTL := secondsToEndOfMonth(now)

	result, err := r.script.Run(ctx, r.client, []string{dailyKey, monthlyKey},
		tokens, r.config.DailyLimit, r.config.MonthlyLimit,
		now.UnixMilli(), dailyTTL, monthlyTTL).Slice()

	if err != nil {
		return false, 0, err
	}

	allowed := result[0].(int64) == 1
	remainingDaily := int(result[1].(int64))

	return allowed, remainingDaily, nil
}

// DecrementBudget decrements the budget (used when request fails or needs rollback)
func (r *RedisBudgetChecker) DecrementBudget(ctx context.Context, key string, tokens int) (int, error) {
	dailyKey := fmt.Sprintf("budget:daily:%s", key)
	monthlyKey := fmt.Sprintf("budget:monthly:%s", key)

	// Use Lua script to atomically decrement both counters
	script := `
local daily_key = KEYS[1]
local monthly_key = KEYS[2]
local tokens = tonumber(ARGV[1])

local daily_used = tonumber(redis.call('GET', daily_key) or '0')
local monthly_used = tonumber(redis.call('GET', monthly_key) or '0')

local new_daily = math.max(0, daily_used - tokens)
local new_monthly = math.max(0, monthly_used - tokens)

redis.call('SET', daily_key, new_daily)
redis.call('SET', monthly_key, new_monthly)

return {new_daily, new_monthly}
`

	_, err := r.client.Eval(ctx, script, []string{dailyKey, monthlyKey}, tokens).Slice()
	if err != nil {
		return 0, err
	}

	// Return remaining daily tokens
	remaining, _, err := r.GetBudget(ctx, key)
	return remaining, err
}

// GetBudget returns current used and limit for a key
func (r *RedisBudgetChecker) GetBudget(ctx context.Context, key string) (used, limit int, err error) {
	dailyKey := fmt.Sprintf("budget:daily:%s", key)

	used, err = r.client.Get(ctx, dailyKey).Int()
	if err == redis.Nil {
		return 0, r.config.DailyLimit, nil
	}
	if err != nil {
		return 0, 0, err
	}

	return used, r.config.DailyLimit, nil
}

// ResetBudget resets the budget for a key
func (r *RedisBudgetChecker) ResetBudget(ctx context.Context, key string) error {
	dailyKey := fmt.Sprintf("budget:daily:%s", key)
	monthlyKey := fmt.Sprintf("budget:monthly:%s", key)

	return r.client.Del(ctx, dailyKey, monthlyKey).Err()
}

// NoOpBudgetChecker is a budget checker that allows all requests
type NoOpBudgetChecker struct{}

// CheckBudget always returns allowed
func (n *NoOpBudgetChecker) CheckBudget(ctx context.Context, key string, tokens int) (bool, int, error) {
	return true, 0, nil
}

// DecrementBudget is a no-op
func (n *NoOpBudgetChecker) DecrementBudget(ctx context.Context, key string, tokens int) (int, error) {
	return 0, nil
}

// GetBudget returns no usage
func (n *NoOpBudgetChecker) GetBudget(ctx context.Context, key string) (used, limit int, err error) {
	return 0, 0, nil
}

// ResetBudget is a no-op
func (n *NoOpBudgetChecker) ResetBudget(ctx context.Context, key string) error {
	return nil
}

// BudgetMiddleware creates an HTTP middleware that enforces token budgets
func BudgetMiddleware(checker BudgetChecker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := GetPrincipalFromContext(r.Context())
		if principal == nil {
			next.ServeHTTP(w, r)
			return
		}

		estimatedTokens := estimateRequestTokens(r)
		allowed, remaining, err := checker.CheckBudget(r.Context(), principal.APIKeyID, estimatedTokens)
		if err != nil {
			http.Error(w, "Budget service unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("X-Budget-Remaining", fmt.Sprintf("%d", remaining))
		if !allowed {
			http.Error(w, "Token budget exceeded", http.StatusPaymentRequired)
			return
		}

		// Store budget checker in request context for post-request decrement
		ctx := context.WithValue(r.Context(), budgetCheckerContextKey, checker)
		ctx = context.WithValue(ctx, budgetTokensContextKey, estimatedTokens)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// budgetCheckerContextKey is the context key for the budget checker
type budgetContextKey string

const (
	budgetCheckerContextKey budgetContextKey = "budget_checker"
	budgetTokensContextKey  budgetContextKey = "budget_tokens"
)

// GetBudgetInfoFromContext retrieves budget-related info from context after request
func GetBudgetInfoFromContext(ctx context.Context) (BudgetChecker, int, bool) {
	checker, ok := ctx.Value(budgetCheckerContextKey).(BudgetChecker)
	if !ok {
		return nil, 0, false
	}
	tokens, ok := ctx.Value(budgetTokensContextKey).(int)
	if !ok {
		return checker, 0, true
	}
	return checker, tokens, true
}

// DecrementBudgetFromContext decrements budget after request completion
func DecrementBudgetFromContext(ctx context.Context, key string) error {
	checker, tokens, ok := GetBudgetInfoFromContext(ctx)
	if !ok || checker == nil {
		return nil
	}
	_, err := checker.DecrementBudget(ctx, key, tokens)
	return err
}

func estimateRequestTokens(r *http.Request) int {
	if r.ContentLength > 0 {
		return int(r.ContentLength) / 4
	}
	return 100
}

// secondsToEndOfDay calculates seconds until end of day UTC
func secondsToEndOfDay(t time.Time) int64 {
	endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
	return int64(endOfDay.Sub(t).Seconds())
}

// secondsToEndOfMonth calculates seconds until end of month UTC
func secondsToEndOfMonth(t time.Time) int64 {
	endOfMonth := time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 0, time.UTC)
	return int64(endOfMonth.Sub(t).Seconds())
}
