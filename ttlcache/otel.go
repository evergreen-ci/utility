package ttlcache

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.GetTracerProvider().Tracer("github.com/evergreen-ci/utility/cache")

const (
	ttlCacheAttribute = "evergreen.cache.ttl"
)

var (
	ttlCacheNameAttribute  = fmt.Sprintf("%s.name", ttlCacheAttribute)
	ttlCacheIDAttribute    = fmt.Sprintf("%s.id", ttlCacheAttribute)
	ttlCacheFoundAttribute = fmt.Sprintf("%s.found", ttlCacheAttribute)
)

// WithOtel wraps a cache and adds OpenTelemetry tracing to it.
// Since this tracks the id, do not use this if the id is sensitive.
// This can be safely used with sensitive values.
func WithOtel[T any](cache Cache[T], name string) *OtelTTLCache[T] {
	return &OtelTTLCache[T]{cache: cache, name: name}
}

type OtelTTLCache[T any] struct {
	cache Cache[T]
	name  string
}

func (c *OtelTTLCache[T]) Get(ctx context.Context, id string, minimumLifetime time.Duration) (T, bool) {
	ctx, span := tracer.Start(ctx, "cache.Get")
	defer span.End()

	value, ok := c.cache.Get(ctx, id, minimumLifetime)

	span.SetAttributes(
		attribute.String(ttlCacheNameAttribute, c.name),
		attribute.String(ttlCacheIDAttribute, id),
		attribute.Bool(ttlCacheFoundAttribute, ok),
	)

	return value, ok
}

func (c *OtelTTLCache[T]) Put(ctx context.Context, id string, value T, expiresAt time.Time) {
	ctx, span := tracer.Start(ctx, "cache.Put")
	defer span.End()

	span.SetAttributes(
		attribute.String(ttlCacheNameAttribute, c.name),
		attribute.String(ttlCacheIDAttribute, id),
	)

	c.cache.Put(ctx, id, value, expiresAt)
}
