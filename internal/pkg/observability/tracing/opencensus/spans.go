package opencensus

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"go.opencensus.io/trace"
)

// StartSpan starts a span.
func StartSpan(ctx context.Context) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, tracing.GetCallerName())
}
