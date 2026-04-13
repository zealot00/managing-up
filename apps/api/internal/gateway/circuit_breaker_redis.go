package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// circuitBreakerLua handles state transitions:
// - closed -> open (after failureThreshold failures)
// - open -> half_open (after timeout)
// Returns: state (0=closed, 1=open, 2=half_open)
const circuitBreakerLua = `
local stateKey = KEYS[1]
local countKey = KEYS[2]
local failureThreshold = tonumber(ARGV[1])
local timeout = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local state = redis.call('GET', stateKey)

if state == false then
    -- Circuit is closed, initialize
    redis.call('SET', stateKey, 'closed', 'PX', timeout)
    redis.call('SET', countKey, 1, 'PX', timeout)
    return 0
end

if state == 'closed' then
    local count = redis.call('INCR', countKey)
    redis.call('PEXPIRE', stateKey, timeout)
    if count >= failureThreshold then
        redis.call('SET', stateKey, 'open', 'PX', timeout)
        redis.call('SET', countKey .. ':failures', 1, 'PX', timeout)
        return 1
    end
    redis.call('PEXPIRE', countKey, timeout)
    return 0
end

if state == 'open' then
    -- Check if timeout has passed, transition to half_open
    redis.call('SET', stateKey, 'half_open', 'PX', timeout)
    redis.call('SET', countKey, 1, 'PX', timeout)
    return 2
end

if state == 'half_open' then
    local count = redis.call('INCR', countKey)
    redis.call('PEXPIRE', countKey, timeout)
    return 2
end

return 0
`

// circuitBreakerSuccessLua records success in half_open state, potentially closing the circuit
// Returns: state (0=closed, 1=open, 2=half_open)
const circuitBreakerSuccessLua = `
local stateKey = KEYS[1]
local countKey = KEYS[2]
local successThreshold = tonumber(ARGV[1])
local timeout = tonumber(ARGV[2])

local state = redis.call('GET', stateKey)

if state == 'half_open' then
    local count = redis.call('INCR', countKey)
    if count >= successThreshold then
        redis.call('SET', stateKey, 'closed', 'PX', timeout)
        redis.call('DEL', countKey, countKey .. ':failures', countKey .. ':successes')
        return 0
    end
    redis.call('PEXPIRE', countKey, timeout)
    return 2
end

if state == 'closed' then
    return 0
end

return 1
`

// circuitBreakerFailureLua records failure, potentially opening the circuit
// Returns: state (0=closed, 1=open, 2=half_open)
const circuitBreakerFailureLua = `
local stateKey = KEYS[1]
local countKey = KEYS[2]
local failureThreshold = tonumber(ARGV[1])
local timeout = tonumber(ARGV[2])

local state = redis.call('GET', stateKey)

if state == 'half_open' then
    redis.call('SET', stateKey, 'open', 'PX', timeout)
    redis.call('DEL', countKey, countKey .. ':failures', countKey .. ':successes')
    return 1
end

if state == 'closed' then
    local count = redis.call('INCR', countKey .. ':failures')
    if count >= failureThreshold then
        redis.call('SET', stateKey, 'open', 'PX', timeout)
        return 1
    end
    return 0
end

return 1
`

// RedisCircuitBreaker implements a distributed circuit breaker using Redis
type RedisCircuitBreaker struct {
	client           *redis.Client
	keyPrefix        string
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	stateScript      *redis.Script
	successScript    *redis.Script
	failureScript    *redis.Script
}

// NewRedisCircuitBreaker creates a new Redis-backed circuit breaker
func NewRedisCircuitBreaker(
	client *redis.Client,
	keyPrefix string,
	failureThreshold int,
	successThreshold int,
	timeout time.Duration,
) *RedisCircuitBreaker {
	return &RedisCircuitBreaker{
		client:           client,
		keyPrefix:        keyPrefix,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		stateScript:      redis.NewScript(circuitBreakerLua),
		successScript:    redis.NewScript(circuitBreakerSuccessLua),
		failureScript:    redis.NewScript(circuitBreakerFailureLua),
	}
}

// stateKey returns the Redis key for circuit state
func (cb *RedisCircuitBreaker) stateKey(key string) string {
	return fmt.Sprintf("%s:cb:%s:state", cb.keyPrefix, key)
}

// countKey returns the Redis key for operation count
func (cb *RedisCircuitBreaker) countKey(key string) string {
	return fmt.Sprintf("%s:cb:%s:count", cb.keyPrefix, key)
}

// State returns the current state of the circuit breaker for the given key
func (cb *RedisCircuitBreaker) State(ctx context.Context, key string) (CircuitBreakerState, error) {
	state, err := cb.client.Get(ctx, cb.stateKey(key)).Result()
	if err == redis.Nil {
		return CircuitBreakerClosed, nil
	}
	if err != nil {
		return CircuitBreakerClosed, err
	}
	return CircuitBreakerState(state), nil
}

// Allow checks if the circuit breaker allows the operation
// Returns true if the circuit is closed or half_open (allowing requests)
// If Redis is unavailable, fails open (allows request) to avoid blocking all traffic
func (cb *RedisCircuitBreaker) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()

	result, err := cb.stateScript.Run(
		ctx,
		cb.client,
		[]string{cb.stateKey(key), cb.countKey(key)},
		cb.failureThreshold,
		cb.timeout.Milliseconds(),
		now,
	).Int()
	if err != nil {
		return true, err
	}

	return result != 1, nil
}

// RecordSuccess records a successful operation
// In half_open state, enough successes will close the circuit
func (cb *RedisCircuitBreaker) RecordSuccess(ctx context.Context, key string) error {
	_, err := cb.successScript.Run(
		ctx,
		cb.client,
		[]string{cb.stateKey(key), cb.countKey(key)},
		cb.successThreshold,
		cb.timeout.Milliseconds(),
	).Int()
	return err
}

// RecordFailure records a failed operation
// In half_open state, any failure opens the circuit
// In closed state, enough failures open the circuit
func (cb *RedisCircuitBreaker) RecordFailure(ctx context.Context, key string) error {
	_, err := cb.failureScript.Run(
		ctx,
		cb.client,
		[]string{cb.stateKey(key), cb.countKey(key)},
		cb.failureThreshold,
		cb.timeout.Milliseconds(),
	).Int()
	return err
}

// Reset clears the circuit breaker state for the given key
func (cb *RedisCircuitBreaker) Reset(ctx context.Context, key string) error {
	stateKey := cb.stateKey(key)
	countKey := cb.countKey(key)
	return cb.client.Del(ctx, stateKey, countKey, countKey+":failures", countKey+":successes").Err()
}
