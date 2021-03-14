package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

var _ UnitCounter = (*unitCounter)(nil)

type unitCounter struct {
	counter metric.Int64Counter
}

func (c *unitCounter) Increment(ctx context.Context, labels ...attribute.KeyValue) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	//c.meter.RecordBatch(ctx, nil, c.counter.Measurement(1))
	c.counter.Add(ctx, 1, labels...)
}

func (c *unitCounter) IncrementBy(ctx context.Context, val int64, labels ...attribute.KeyValue) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	//c.meter.RecordBatch(ctx, nil, c.counter.Measurement(val))
	c.counter.Add(ctx, val, labels...)
}

func (c *unitCounter) Decrement(ctx context.Context, labels ...attribute.KeyValue) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	//c.meter.RecordBatch(ctx, nil, c.counter.Measurement(-1))
	c.counter.Add(ctx, -1, labels...)
}
