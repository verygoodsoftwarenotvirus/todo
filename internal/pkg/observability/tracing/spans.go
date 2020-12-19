package tracing

import (
	"context"

	"go.opencensus.io/trace"
)

// StartSpan starts a span.
func StartSpan(ctx context.Context) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, GetCallerName())
}
