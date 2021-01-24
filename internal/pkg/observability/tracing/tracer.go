package tracing

import (
	"fmt"

	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupJaeger creates a new trace provider instance and registers it as global trace provider.
func SetupJaeger(endpoint, serviceName string) (func(), error) {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(endpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)

	if err != nil {
		return nil, fmt.Errorf("initializing Jaeger: %w", err)
	}

	return flush, nil
}
