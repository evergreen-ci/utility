package ttlcache

import (
	"context"
	"time"
	"weak"
)

// NewWeakInMemory creates a new thread-safe in-memory ttl cache that uses weak references.
func NewWeakInMemory[T any]() *WeakInMemory[T] {
	return &WeakInMemory[T]{
		cache: NewInMemory[weak.Pointer[T]](),
	}
}

type WeakInMemory[T any] struct {
	cache *InMemoryCache[weak.Pointer[T]]
}

func (w *WeakInMemory[T]) Get(ctx context.Context, id string, minimumLifetime time.Duration) (T, bool) {
	weakVal, found := w.cache.Get(ctx, id, minimumLifetime)
	if !found {
		var zero T
		return zero, false
	}
	val := weakVal.Value()
	if val == nil {
		var zero T
		return zero, false
	}

	return *val, true
}

func (w *WeakInMemory[T]) Put(ctx context.Context, id string, value T, expiresAt time.Time) {
	w.cache.Put(ctx, id, weak.Make(&value), expiresAt)
}
