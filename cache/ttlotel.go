package cache

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.GetTracerProvider().Tracer("github.com/evergreen-ci/utility/cache")

const (
	ttlCacheFoundAttribute = "evergreen.cache.ttl.%s.found"
)

// newTTLOtelCache wraps a cache and adds OpenTelemetry tracing to it.
func newTTLOtelCache[T any](cache TTLCache[T]) TTLCache[T] {
	return &otelTTLCache[T]{cache: cache}
}

type otelTTLCache[T any] struct {
	cache TTLCache[T]
}

func (c *otelTTLCache[T]) Get(ctx context.Context, id string, minimumLifetime time.Duration) (T, bool) {
	ctx, span := tracer.Start(ctx, "cache.Get")
	defer span.End()

	value, ok := c.cache.Get(ctx, id, minimumLifetime)

	span.SetAttributes(
		attribute.Bool(fmt.Sprintf(ttlCacheFoundAttribute, c.name()), ok),
	)

	return value, ok
}

func (c *otelTTLCache[T]) Put(ctx context.Context, id string, value T, expiresAt time.Time) {
	ctx, span := tracer.Start(ctx, "cache.Put")
	defer span.End()

	c.cache.Put(ctx, id, value, expiresAt)
}

func (c *otelTTLCache[T]) name() string {
	return c.cache.name()
}
