package ttlcache

import (
	"context"
	"sync"
	"time"
)

// NewInMemory creates a new thread-safe in-memory ttl cache.
func NewInMemory[T any]() *InMemoryCache[T] {
	return &InMemoryCache[T]{
		mu:    sync.RWMutex{},
		cache: make(map[string]ttlValue[T]),
	}
}

type InMemoryCache[T any] struct {
	mu    sync.RWMutex
	cache map[string]ttlValue[T]
}

func (c *InMemoryCache[T]) Get(_ context.Context, id string, minimumLifetime time.Duration) (T, bool) {
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

func (c *InMemoryCache[T]) Put(_ context.Context, id string, value T, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[id] = ttlValue[T]{
		value:     value,
		expiresAt: expiresAt,
	}
}

func (c *InMemoryCache[T]) Delete(_ context.Context, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, id)
}
