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
