package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter provides rate limiting functionality using Redis
type Limiter struct {
	redis             *redis.Client
	defaultLimitPerMin int
	windowDuration     time.Duration
}

// NewLimiter creates a new rate limiter
func NewLimiter(redisClient *redis.Client, defaultLimit int) *Limiter {
	return &Limiter{
		redis:             redisClient,
		defaultLimitPerMin: defaultLimit,
		windowDuration:     time.Minute,
	}
}

// CheckGlobalLimit checks if global rate limit is exceeded (deprecated - kept for compatibility)
// Note: Now uses default limit as there's no separate global limit
func (l *Limiter) CheckGlobalLimit(ctx context.Context) (bool, error) {
	key := "ratelimit:global"
	return l.checkLimit(ctx, key, l.defaultLimitPerMin)
}

// CheckUserLimit checks if user-specific rate limit is exceeded
// If customLimit is nil, uses the default limit
func (l *Limiter) CheckUserLimit(ctx context.Context, userID string, customLimit *int) (bool, error) {
	key := fmt.Sprintf("ratelimit:user:%s", userID)

	limit := l.defaultLimitPerMin
	if customLimit != nil && *customLimit > 0 {
		limit = *customLimit
	}

	return l.checkLimit(ctx, key, limit)
}

// checkLimit implements sliding window rate limiting
func (l *Limiter) checkLimit(ctx context.Context, key string, limit int) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-l.windowDuration)

	pipe := l.redis.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count current entries in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiration
	pipe.Expire(ctx, key, l.windowDuration*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit check: %w", err)
	}

	count := countCmd.Val()

	// If count >= limit, rate limit exceeded
	return count >= int64(limit), nil
}

// GetRemainingRequests returns the number of remaining requests for a user
// If customLimit is nil, uses the default limit
func (l *Limiter) GetRemainingRequests(ctx context.Context, userID string, customLimit *int) (int, error) {
	key := fmt.Sprintf("ratelimit:user:%s", userID)

	limit := l.defaultLimitPerMin
	if customLimit != nil && *customLimit > 0 {
		limit = *customLimit
	}

	now := time.Now()
	windowStart := now.Add(-l.windowDuration)

	// Count current entries in window
	count, err := l.redis.ZCount(ctx, key, fmt.Sprintf("%d", windowStart.UnixNano()), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// ResetUserLimit resets rate limit for a specific user
func (l *Limiter) ResetUserLimit(ctx context.Context, userID string) error {
	key := fmt.Sprintf("ratelimit:user:%s", userID)
	return l.redis.Del(ctx, key).Err()
}

// ResetGlobalLimit resets global rate limit
func (l *Limiter) ResetGlobalLimit(ctx context.Context) error {
	key := "ratelimit:global"
	return l.redis.Del(ctx, key).Err()
}

// GetUserRequestCount gets the current request count for a user
func (l *Limiter) GetUserRequestCount(ctx context.Context, userID string) (int64, error) {
	key := fmt.Sprintf("ratelimit:user:%s", userID)
	now := time.Now()
	windowStart := now.Add(-l.windowDuration)

	count, err := l.redis.ZCount(ctx, key, fmt.Sprintf("%d", windowStart.UnixNano()), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	return count, nil
}

// IncrementCounter increments a simple counter (for statistics)
func (l *Limiter) IncrementCounter(ctx context.Context, key string, expiration time.Duration) error {
	pipe := l.redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	return err
}

// GetCounter gets a counter value
func (l *Limiter) GetCounter(ctx context.Context, key string) (int64, error) {
	val, err := l.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}
