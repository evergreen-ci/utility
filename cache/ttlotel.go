package cache

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("cache")

const (
	// The unformatted string is the name of the cache.
	ttlCache = "evergreen.cache.ttl.%s"
)

var (
	ttlCacheFoundAttribute       = fmt.Sprintf("%s.found", ttlCache)
	ttlCacheTimeTakenMSAttribute = fmt.Sprintf("%s.time_taken_ms", ttlCache)
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

	before := time.Now()
	value, ok := c.cache.Get(ctx, id, minimumLifetime)

	span.SetAttributes(
		attribute.Bool(fmt.Sprintf(ttlCacheFoundAttribute, c.name()), ok),
		attribute.Float64(fmt.Sprintf(ttlCacheTimeTakenMSAttribute, c.name()), float64(time.Since(before).Milliseconds())),
	)

	return value, ok
}

func (c *otelTTLCache[T]) Put(ctx context.Context, id string, value T, expiresAt time.Time) {
	ctx, span := tracer.Start(ctx, "cache.Put")
	defer span.End()

	before := time.Now()
	c.cache.Put(ctx, id, value, expiresAt)

	span.SetAttributes(
		attribute.Float64(fmt.Sprintf(ttlCacheTimeTakenMSAttribute, c.name()), float64(time.Since(before).Milliseconds())),
	)
}

func (c *otelTTLCache[T]) name() string {
	return c.cache.name()
}
