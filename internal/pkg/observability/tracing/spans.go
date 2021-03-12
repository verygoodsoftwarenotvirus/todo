package tracing

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StartCustomSpan starts an anonymous custom span.
func StartCustomSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	return otel.Tracer("_anon_").Start(ctx, name)
}

// StartSpan starts an anonymous span.
func StartSpan(ctx context.Context) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	return otel.Tracer("_anon_").Start(ctx, GetCallerName())
}

// Tracer describes a tracer.
type Tracer interface {
	StartSpan(ctx context.Context) (context.Context, trace.Span)
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
	if ctx == nil {
		ctx = context.Background()
	}

	return t.tracer.Start(ctx, GetCallerName())
}

var uriIDReplacementRegex = regexp.MustCompile(`\/\d+`)

// FormatSpan formats a span
func FormatSpan(operation string, req *http.Request) string {
	spanURI := fmt.Sprintf(uriIDReplacementRegex.ReplaceAllString(req.URL.Path, "/<id>"))
	return fmt.Sprintf("%s %s: %s", req.Method, spanURI, operation)
}
