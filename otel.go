package utility

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

type otelAttributeKey int

const otelAttributeContextKey otelAttributeKey = iota

type attributeSpanProcessor struct{}

// NewAttributeSpanProcessor returns a span processor that adds all the attributes added to the context
// to every span created from that context and its children.
func NewAttributeSpanProcessor() trace.SpanProcessor {
	return &attributeSpanProcessor{}
}

// OnStart adds the attributes stored in the context to a span when the span is started.
// It is called by the otel SDK.
func (processor *attributeSpanProcessor) OnStart(ctx context.Context, span trace.ReadWriteSpan) {
	span.SetAttributes(attributesFromContext(ctx)...)
}

// OnEnd is a noop to satisfy the SpanProcessor interface.
func (processor *attributeSpanProcessor) OnEnd(s trace.ReadOnlySpan) {}

// Shutdown is a noop to satisfy the SpanProcessor interface.
func (processor *attributeSpanProcessor) Shutdown(context.Context) error { return nil }

// ForceFlush is a noop to satisfy the SpanProcessor interface.
func (processor *attributeSpanProcessor) ForceFlush(context.Context) error { return nil }

// ContextWithAttributes returns a child of the ctx containing the attributes. All spans
// created with the returned context and its children will have the attributes set.
func ContextWithAttributes(ctx context.Context, attributes []attribute.KeyValue) context.Context {
	return context.WithValue(ctx, otelAttributeContextKey, attributes)
}

func attributesFromContext(ctx context.Context) []attribute.KeyValue {
	attributesIface := ctx.Value(otelAttributeContextKey)
	attributes, ok := attributesIface.([]attribute.KeyValue)
	if !ok {
		return nil
	}
	return attributes
}
