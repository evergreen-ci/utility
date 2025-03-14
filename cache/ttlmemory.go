package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NewTTLInMemory creates a new thread-safe in-memory ttl cache.
// This is most useful when storing tokens that will be handed off to
// other services.
func NewTTLInMemory[T any](name string) TTLCache[T] {
	return &ttlInMemoryCache[T]{
		_name: name,
		mu:    sync.RWMutex{},
		cache: make(map[string]ttlValue[T]),
	}
}

type ttlInMemoryCache[T any] struct {
	_name string

	mu    sync.RWMutex
	cache map[string]ttlValue[T]
}

func (c *ttlInMemoryCache[T]) Get(_ context.Context, id string, minimumLifetime time.Duration) (T, bool) {
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

func (c *ttlInMemoryCache[T]) Put(_ context.Context, id string, value T, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[id] = ttlValue[T]{
		value:     value,
		expiresAt: expiresAt,
	}
}

func (c *ttlInMemoryCache[T]) name() string {
	return fmt.Sprintf("in-memory-ttl.%s", c._name)
}
