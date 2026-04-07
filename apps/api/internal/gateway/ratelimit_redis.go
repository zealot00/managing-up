package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisRateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local current = redis.call('GET', key)
if current == false then
    redis.call('SET', key, 1, 'PX', window)
    redis.call('SET', key .. ':reset', now + window, 'PX', window)
    return 1
end

current = tonumber(current)
if current >= limit then
    return 0
end

local new_count = redis.call('INCR', key)
if new_count == 1 then
    redis.call('PEXPIRE', key, window)
    redis.call('SET', key .. ':reset', now + window, 'PX', window)
end

return 1
`

type RedisRateLimiter struct {
	client    *redis.Client
	keyPrefix string
	limit     int
	window    time.Duration
	script    *redis.Script
}

func NewRedisRateLimiter(client *redis.Client, keyPrefix string, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:    client,
		keyPrefix: keyPrefix,
		limit:     limit,
		window:    window,
		script:    redis.NewScript(redisRateLimitScript),
	}
}

func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)
	now := time.Now().UnixMilli()

	result, err := rl.script.Run(ctx, rl.client, []string{fullKey}, rl.limit, rl.window.Milliseconds(), now).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (rl *RedisRateLimiter) Remaining(ctx context.Context, key string) (int, error) {
	fullKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)

	current, err := rl.client.Get(ctx, fullKey).Int()
	if err == redis.Nil {
		return rl.limit, nil
	}
	if err != nil {
		return 0, err
	}

	remaining := rl.limit - current
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

func (rl *RedisRateLimiter) ResetAt(ctx context.Context, key string) (time.Time, error) {
	fullKey := fmt.Sprintf("%s:%s:reset", rl.keyPrefix, key)

	resetAtMs, err := rl.client.Get(ctx, fullKey).Int64()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	return time.UnixMilli(resetAtMs), nil
}

func (rl *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)
	resetKey := fmt.Sprintf("%s:%s:reset", rl.keyPrefix, key)

	return rl.client.Del(ctx, fullKey, resetKey).Err()
}

type RedisRateLimiterFactory struct {
	Client *redis.Client
}

func (f *RedisRateLimiterFactory) Create(keyPrefix string, limit int, window time.Duration) RateLimiter {
	return NewRedisRateLimiter(f.Client, keyPrefix, limit, window)
}

func (rl *RedisRateLimiter) Limit() int {
	return rl.limit
}
