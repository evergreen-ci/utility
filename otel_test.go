package utility

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestContextWithAttributes(t *testing.T) {
	t.Run("HasNoInitialAttributes", func(t *testing.T) {
		ctx := context.Background()
		assert.Empty(t, attributesFromContext(ctx))
	})
	t.Run("HasNoAttributesWhenAddingNil", func(t *testing.T) {
		ctx := ContextWithAttributes(context.Background(), nil)
		assert.Empty(t, attributesFromContext(ctx))
	})
	t.Run("HasAttributesAfterSetting", func(t *testing.T) {
		attrs := []attribute.KeyValue{
			attribute.String("key1", "value1"),
			attribute.Int("key2", 42),
		}
		ctx := ContextWithAttributes(context.Background(), attrs)
		assert.Equal(t, attrs, attributesFromContext(ctx))

		childCtx := context.WithValue(ctx, "unrelatedContextKey", "unrelatedContextValue")
		assert.Equal(t, attrs, attributesFromContext(childCtx), "child context should inherit span attributes from parent")
	})
	t.Run("OverwritesExistingAttributes", func(t *testing.T) {
		oldAttrs := []attribute.KeyValue{
			attribute.String("oldKey", "oldValue"),
		}
		ctx := ContextWithAttributes(context.Background(), oldAttrs)
		assert.Equal(t, oldAttrs, attributesFromContext(ctx))

		newAttrs := []attribute.KeyValue{
			attribute.String("newKey", "newValue"),
		}
		newCtx := ContextWithAttributes(ctx, newAttrs)
		assert.Equal(t, newAttrs, attributesFromContext(newCtx), "new context should have overwritten old attributes")
		assert.Equal(t, oldAttrs, attributesFromContext(ctx), "original context attributes should remain unchanged")
	})
}

func TestAppendAttributesToContext(t *testing.T) {
	t.Run("AppendsToInitiallyEmptyContext", func(t *testing.T) {
		attrs := []attribute.KeyValue{
			attribute.String("key1", "value1"),
			attribute.Int("key2", 42),
		}
		ctx := ContextWithAppendedAttributes(context.Background(), attrs)
		assert.Equal(t, attrs, attributesFromContext(ctx))

		childCtx := context.WithValue(ctx, "unrelatedContextKey", "unrelatedContextValue")
		assert.Equal(t, attrs, attributesFromContext(childCtx), "child context should inherit span attributes from parent")
	})
	t.Run("AppendsToContextWithPreexistingAttributes", func(t *testing.T) {
		oldAttrs := []attribute.KeyValue{
			attribute.String("oldKey", "oldValue"),
		}
		oldCtx := ContextWithAppendedAttributes(context.Background(), oldAttrs)
		assert.Equal(t, oldAttrs, attributesFromContext(oldCtx))

		newAttrs := []attribute.KeyValue{
			attribute.String("newKey", "newValue"),
		}
		newCtx := ContextWithAppendedAttributes(oldCtx, newAttrs)
		expectedAttrs := append(oldAttrs, newAttrs...)
		assert.Equal(t, expectedAttrs, attributesFromContext(newCtx), "new context should have both previous and newly-added attributes")
		assert.Equal(t, oldAttrs, attributesFromContext(oldCtx), "original context attributes should remain unchanged")
	})
}
