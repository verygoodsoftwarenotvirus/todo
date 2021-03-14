package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
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
