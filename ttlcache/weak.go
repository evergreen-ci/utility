package cache

import (
	"context"
	"time"
	"weak"
)

func NewWeakInMemory[T any]() *WeakInMemory[T] {
	return &WeakInMemory[T]{
		cache: NewTTLInMemory[weak.Pointer[T]](),
	}
}

type WeakInMemory[T any] struct {
	cache *TTLInMemoryCache[weak.Pointer[T]]
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
