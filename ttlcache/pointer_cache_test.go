package ttlcache

import (
	"context"
	"time"
)

// convertPointerCacheToCache creates a wrapper around a PointerCache that
// adapts it to the Cache interface.
// This isn't intended to be used in production code as it will
// allocate a new value on every Get call. It is only intended to be used
// for testing purposes.
// In production code, use the PointerCache directly.
func convertPointerCacheToCache[T any](ptrCache PointerCache[T]) Cache[T] {
	return &pointerToValueCache[T]{cache: ptrCache}
}

type pointerToValueCache[T any] struct {
	cache PointerCache[T]
}

func (c *pointerToValueCache[T]) Get(ctx context.Context, id string, minimumLifetime time.Duration) (T, bool) {
	ptr, ok := c.cache.Get(ctx, id, minimumLifetime)
	if !ok || ptr == nil {
		var zero T
		return zero, false
	}
	return *ptr, true
}

func (c *pointerToValueCache[T]) Put(ctx context.Context, id string, value T, expiresAt time.Time) {
	c.cache.Put(ctx, id, &value, expiresAt)
}

func (c *pointerToValueCache[T]) Delete(ctx context.Context, id string) {
	c.cache.Delete(ctx, id)
}
