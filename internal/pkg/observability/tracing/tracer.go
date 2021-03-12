package tracing

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
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
func (cfg *Config) SetupJaeger() (func(), error) {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(cfg.Jaeger.CollectorEndpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: cfg.Jaeger.ServiceName,
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithBatchMaxCount(10), // set this to 1 to report every span immediately.
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.TraceIDRatioBased(cfg.SpanCollectionProbability)}),
	)

	if err != nil {
		return nil, fmt.Errorf("initializing Jaeger: %w", err)
	}

	return flush, nil
}
