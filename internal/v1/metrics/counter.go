package metrics

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// Counter counts things
type Counter interface {
	Increment()
	IncrementBy(val uint64)
	Decrement()
}

// opencensusCounter is a Counter that interfaces with opencensus
type opencensusCounter struct {
	name        string
	actualCount uint64
	count       *stats.Int64Measure
	counter     *view.View
}

// Increment satisfies our Counter interface
func (c *opencensusCounter) Increment(ctx context.Context) {
	atomic.AddUint64(&c.actualCount, 1)
	stats.Record(ctx, c.count.M(1))
}

// IncrementBy satisfies our Counter interface
func (c *opencensusCounter) IncrementBy(ctx context.Context, val uint64) {
	atomic.AddUint64(&c.actualCount, val)
	stats.Record(ctx, c.count.M(int64(val)))
}

// Decrement satisfies our Counter interface
func (c *opencensusCounter) Decrement(ctx context.Context) {
	atomic.AddUint64(&c.actualCount, ^uint64(0))
	stats.Record(ctx, c.count.M(-1))
}

// ProvideUnitCounterProvider provides UnitCounter providers
func ProvideUnitCounterProvider() UnitCounterProvider {
	return ProvideUnitCounter
}

// ProvideUnitCounter provides a new counter
func ProvideUnitCounter(counterName CounterName, description string) (UnitCounter, error) {
	name := fmt.Sprintf("%s_count", string(counterName))
	// Counts/groups the lengths of lines read in.
	count := stats.Int64(name, "", "By")

	countView := &view.View{
		Name:        name,
		Description: description,
		Measure:     count,
		Aggregation: view.Count(),
	}

	if err := view.Register(countView); err != nil {
		return nil, fmt.Errorf("failed to register views: %w", err)
	}

	return &opencensusCounter{
		name:    name,
		count:   count,
		counter: countView,
	}, nil
}
