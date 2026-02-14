package ratelimit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// MemoryLimiter is an in-memory rate limiter (per key). Suitable for single instance.
type MemoryLimiter struct {
	limit  rate.Limit
	limiters map[string]*rate.Limiter
	mu     sync.Mutex
}

// NewMemoryLimiter creates a limiter allowing limitPerSec events per second per key.
func NewMemoryLimiter(limitPerSec int) *MemoryLimiter {
	return &MemoryLimiter{
		limit:   rate.Limit(limitPerSec),
		limiters: make(map[string]*rate.Limiter),
	}
}

func (l *MemoryLimiter) getLimiter(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[key]; ok {
		return lim
	}
	burst := int(l.limit) * 2
	if burst < 1 {
		burst = 1
	}
	lim := rate.NewLimiter(l.limit, burst)
	l.limiters[key] = lim
	return lim
}

func (l *MemoryLimiter) Allow(ctx context.Context, key string) (bool, error) {
	lim := l.getLimiter(key)
	return lim.Allow(), nil
}

func (l *MemoryLimiter) Wait(ctx context.Context, key string) error {
	lim := l.getLimiter(key)
	return lim.Wait(ctx)
}
