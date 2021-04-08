package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

type tracingErrorHandler struct {
	logger logging.Logger
}

func (h tracingErrorHandler) Handle(err error) {
	h.logger.Error(err, "tracer reported issue")
}

func init() {
	otel.SetErrorHandler(tracingErrorHandler{logger: logging.NewNonOperationalLogger().WithName("otel_errors")})
}

// SetupJaeger creates a new trace provider instance and registers it as global trace provider.
func (c *Config) SetupJaeger() (func(), error) {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(c.Jaeger.CollectorEndpoint),
		jaeger.WithProcessFromEnv(),
		jaeger.WithSDKOptions(
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.SpanCollectionProbability)),
			sdktrace.WithResource(resource.NewWithAttributes(
				attribute.String("exporter", "jaeger"),
			)),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("initializing Jaeger: %w", err)
	}

	return flush, nil
}

// Tracer describes a tracer.
type Tracer interface {
	StartSpan(ctx context.Context) (context.Context, Span)
	StartCustomSpan(ctx context.Context, name string) (context.Context, Span)
}

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
	if ctx == nil {
		ctx = context.Background()
	}

	return t.tracer.Start(ctx, GetCallerName())
}

// StartCustomSpan wraps tracer.Start.
func (t *otelSpanManager) StartCustomSpan(ctx context.Context, name string) (context.Context, Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	return t.tracer.Start(ctx, name)
}
