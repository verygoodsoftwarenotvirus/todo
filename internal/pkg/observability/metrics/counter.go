package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

type unitCounter struct {
	counter metric.Int64UpDownCounter
}

func (c *unitCounter) Increment(ctx context.Context) {
	c.counter.Add(ctx, 1)
}

func (c *unitCounter) IncrementBy(ctx context.Context, val int64) {
	c.counter.Add(ctx, val)
}

func (c *unitCounter) Decrement(ctx context.Context) {
	c.counter.Add(ctx, -1)
}
