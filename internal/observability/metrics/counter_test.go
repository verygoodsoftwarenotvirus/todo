package metrics

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
)

func Test_unitCounter_Decrement(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		initiatePrometheusExporter()
		meterProvider := prometheusExporter.MeterProvider()
		mustMeter := metric.Must(meterProvider.Meter(defaultNamespace, metric.WithInstrumentationVersion(instrumentationVersion)))

		ctx := context.Background()
		uc := &unitCounter{
			counter: mustMeter.NewInt64Counter(
				t.Name(),
				metric.WithUnit(unit.Dimensionless),
			),
		}

		uc.Decrement(ctx)
	})
}

func Test_unitCounter_Increment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		initiatePrometheusExporter()
		meterProvider := prometheusExporter.MeterProvider()
		mustMeter := metric.Must(meterProvider.Meter(defaultNamespace, metric.WithInstrumentationVersion(instrumentationVersion)))

		ctx := context.Background()
		uc := &unitCounter{
			counter: mustMeter.NewInt64Counter(
				t.Name(),
				metric.WithUnit(unit.Dimensionless),
			),
		}

		uc.Increment(ctx)
	})
}

func Test_unitCounter_IncrementBy(T *testing.T) {
	T.Parallel()

	initiatePrometheusExporter()
	meterProvider := prometheusExporter.MeterProvider()
	mustMeter := metric.Must(meterProvider.Meter(defaultNamespace, metric.WithInstrumentationVersion(instrumentationVersion)))

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		uc := &unitCounter{
			counter: mustMeter.NewInt64Counter(
				t.Name(),
				metric.WithUnit(unit.Dimensionless),
			),
		}

		uc.IncrementBy(ctx, 123)
	})
}
