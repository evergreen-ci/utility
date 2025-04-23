package ttlcache

import (
	"context"
	"time"
)

// PointerCache holds pointers in a cache with a time-to-live.
type PointerCache[T any] interface {
	// Get gets the value with id with at least the minimum lifetime remaining.
	Get(ctx context.Context, id string, minimumLifetime time.Duration) (*T, bool)
	// Put adds a value to the cache with the given expiration time.
	Put(ctx context.Context, id string, value *T, expiresAt time.Time)
	// Delete removes the value with id from the cache. This is typically used
	// to clean up expired values. It will no-op if the id is not found.
	Delete(ctx context.Context, id string)
}

// convertPointerCacheToCache creates a wrapper around a PointerCache that
// adapts it to the Cache interface.
// This isn't intended to be used in production code as it will
// allocate a new value on every Get call. It is only intended to be used
// for testing purposes.
// In production code, use the PointerCache directly.
func convertPointerCacheToCache[T any](ptrCache PointerCache[T]) Cache[T] {
	return &PointerToValueCache[T]{cache: ptrCache}
}

type PointerToValueCache[T any] struct {
	cache PointerCache[T]
}

func (c *PointerToValueCache[T]) Get(ctx context.Context, id string, minimumLifetime time.Duration) (T, bool) {
	ptr, ok := c.cache.Get(ctx, id, minimumLifetime)
	if !ok || ptr == nil {
		var zero T
		return zero, false
	}
	return *ptr, true
}

func (c *PointerToValueCache[T]) Put(ctx context.Context, id string, value T, expiresAt time.Time) {
	c.cache.Put(ctx, id, &value, expiresAt)
}

func (c *PointerToValueCache[T]) Delete(ctx context.Context, id string) {
	c.cache.Delete(ctx, id)
}
