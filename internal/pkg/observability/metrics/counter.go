package metrics

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// opencensusCounter is a Counter that interfaces with opencensus.
type opencensusCounter struct {
	name        string
	actualCount uint64
	measure     *stats.Int64Measure
	v           *view.View
}

func (c *opencensusCounter) subtractFromCount(ctx context.Context, value uint64) {
	atomic.AddUint64(&c.actualCount, ^value+1)
	stats.Record(ctx, c.measure.M(int64(-value)))
}

func (c *opencensusCounter) addToCount(ctx context.Context, value uint64) {
	atomic.AddUint64(&c.actualCount, value)
	stats.Record(ctx, c.measure.M(int64(value)))
}

// Decrement satisfies our Counter interface.
func (c *opencensusCounter) Decrement(ctx context.Context) {
	c.subtractFromCount(ctx, 1)
}

// Increment satisfies our Counter interface.
func (c *opencensusCounter) Increment(ctx context.Context) {
	c.addToCount(ctx, 1)
}

// IncrementBy satisfies our Counter interface.
func (c *opencensusCounter) IncrementBy(ctx context.Context, value uint64) {
	c.addToCount(ctx, value)
}

// ProvideUnitCounter provides a new counter.
func ProvideUnitCounter(counterName CounterName, description string) (UnitCounter, error) {
	name := fmt.Sprintf("%s_count", string(counterName))
	// Counts/groups the lengths of lines read in.
	count := stats.Int64(name, description, "By")

	countView := &view.View{
		Name:        name,
		Description: description,
		Measure:     count,
		Aggregation: view.Count(),
	}

	if err := view.Register(countView); err != nil {
		return nil, fmt.Errorf("failed to register views: %w", err)
	}

	c := &opencensusCounter{
		name:    name,
		measure: count,
		v:       countView,
	}

	return c, nil
}
