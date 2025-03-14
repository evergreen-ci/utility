package cache

import (
	"context"
	"sync"
	"time"
)

// NewTTLInMemory creates a new thread-safe in-memory ttl cache.
func NewTTLInMemory[T any]() *TTLInMemoryCache[T] {
	return &TTLInMemoryCache[T]{
		mu:    sync.RWMutex{},
		cache: make(map[string]ttlValue[T]),
	}
}

type TTLInMemoryCache[T any] struct {
	mu    sync.RWMutex
	cache map[string]ttlValue[T]
}

func (c *TTLInMemoryCache[T]) Get(_ context.Context, id string, minimumLifetime time.Duration) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cachedToken, ok := c.cache[id]
	if !ok {
		var value T
		return value, false
	}
	if time.Until(cachedToken.expiresAt) < minimumLifetime {
		var value T
		return value, false
	}

	return cachedToken.value, true
}

func (c *TTLInMemoryCache[T]) Put(_ context.Context, id string, value T, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[id] = ttlValue[T]{
		value:     value,
		expiresAt: expiresAt,
	}
}
