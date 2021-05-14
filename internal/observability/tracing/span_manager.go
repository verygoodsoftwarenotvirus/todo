package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var _ Tracer = (*otelSpanManager)(nil)

type otelSpanManager struct {
	tracer trace.Tracer
}

// NewTracer creates a Tracer.
func NewTracer(name string) Tracer {
	return &otelSpanManager{
		tracer: otel.Tracer(name),
	}
}

// StartSpan wraps tracer.Start.
func (t *otelSpanManager) StartSpan(ctx context.Context) (context.Context, Span) {
	return t.tracer.Start(ctx, GetCallerName())
}

// StartCustomSpan wraps tracer.Start.
func (t *otelSpanManager) StartCustomSpan(ctx context.Context, name string) (context.Context, Span) {
	return t.tracer.Start(ctx, name)
}
