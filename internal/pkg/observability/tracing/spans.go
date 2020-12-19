package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts an anonymous span.
func StartSpan(ctx context.Context) (context.Context, trace.Span) {
	return otel.Tracer("_anon_").Start(ctx, GetCallerName())
}

// Tracer describes a tracer.
type Tracer interface {
	StartSpan(ctx context.Context) (context.Context, trace.Span)
	StartCustomSpan(ctx context.Context, name string) (context.Context, trace.Span)
}

var _ Tracer = (*otSpanManager)(nil)

type otSpanManager struct {
	tracer trace.Tracer
}

// NewTracer creates a Tracer.
func NewTracer(name string) Tracer {
	return &otSpanManager{
		tracer: otel.Tracer(name),
	}
}

// StartSpan wraps tracer.Start.
func (t *otSpanManager) StartSpan(ctx context.Context) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, GetCallerName())
}

// StartSpan wraps tracer.Start.
func (t *otSpanManager) StartCustomSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name)
}
