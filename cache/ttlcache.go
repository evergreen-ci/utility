package cache

import (
	"context"
	"time"
)

// TTLCache holds items in a cache with a time-to-live.
type TTLCache[T any] interface {
	// Get gets the value with id with at least the minimum lifetime remaining.
	Get(ctx context.Context, id string, minimumLifetime time.Duration) (T, bool)
	// Put adds a value to the cache with the given expiration time.
	Put(ctx context.Context, id string, value T, expiresAt time.Time)
	// name returns the name of the cache.
	name() string
}

// ttlValue is a generic type that holds a value and an expiration time.
type ttlValue[T any] struct {
	value     T
	expiresAt time.Time
}
