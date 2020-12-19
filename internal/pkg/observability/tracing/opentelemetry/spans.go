package opentelemetry

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a span.
func StartSpan(ctx context.Context, tracer trace.Tracer) (context.Context, trace.Span) {
	return tracer.Start(ctx, tracing.GetCallerName())
}
