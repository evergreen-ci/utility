package ttlcache

import (
	"context"
	"time"
	"weak"
)

// NewWeakInMemory creates a new thread-safe in-memory ttl cache that uses weak references.
// Weak references allow the garbage collector to reclaim the memory used by the value
// when there are no strong references to it. This is useful for caching large objects
// that may not be needed for long periods of time, as it allows the memory to be reclaimed
// when the object is no longer needed.
func NewWeakInMemory[T any]() *WeakInMemory[T] {
	return &WeakInMemory[T]{
		cache: NewInMemory[weak.Pointer[T]](),
	}
}

type WeakInMemory[T any] struct {
	cache *InMemoryCache[weak.Pointer[T]]
}

func (w *WeakInMemory[T]) Get(ctx context.Context, id string, minimumLifetime time.Duration) (*T, bool) {
	weakVal, found := w.cache.Get(ctx, id, minimumLifetime)
	if !found {
		return nil, false
	}
	val := weakVal.Value()
	if val == nil {
		// Clean up the cache if the value is nil.
		w.Delete(ctx, id)

		return nil, false
	}

	return val, true
}

func (w *WeakInMemory[T]) Put(ctx context.Context, id string, value *T, expiresAt time.Time) {
	w.cache.Put(ctx, id, weak.Make(value), expiresAt)
}

func (c *WeakInMemory[T]) Delete(ctx context.Context, id string) {
	c.cache.Delete(ctx, id)
}
