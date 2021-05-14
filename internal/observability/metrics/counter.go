package metrics

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"go.opentelemetry.io/otel/metric"
)

var _ UnitCounter = (*unitCounter)(nil)

type unitCounter struct {
	counter metric.Int64Counter
}

func (c *unitCounter) Increment(ctx context.Context) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	c.counter.Add(ctx, 1)
}

func (c *unitCounter) IncrementBy(ctx context.Context, val int64) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	c.counter.Add(ctx, val)
}

func (c *unitCounter) Decrement(ctx context.Context) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	c.counter.Add(ctx, -1)
}
