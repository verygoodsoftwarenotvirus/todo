package tracing

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type errorHandler struct {
	logger logging.Logger
}

func (h errorHandler) Handle(err error) {
	h.logger.Error(err, "tracer reported issue")
}

func init() {
	otel.SetErrorHandler(errorHandler{logger: logging.NewNonOperationalLogger().WithName("otel_errors")})
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
