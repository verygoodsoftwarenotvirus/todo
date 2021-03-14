package tracing

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type tracingErrorHandler struct{}

func (h tracingErrorHandler) Handle(err error) {
	logger := zerolog.NewLogger()

	logger.Error(err, "tracer reported issue")
}

func init() {
	otel.SetErrorHandler(tracingErrorHandler{})
}

// SetupJaeger creates a new trace provider instance and registers it as global trace provider.
func (c *Config) SetupJaeger() (func(), error) {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(c.Jaeger.CollectorEndpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: c.Jaeger.ServiceName,
			Tags: []attribute.KeyValue{
				attribute.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.TraceIDRatioBased(c.SpanCollectionProbability)}),
	)

	if err != nil {
		return nil, fmt.Errorf("initializing Jaeger: %w", err)
	}

	return flush, nil
}
