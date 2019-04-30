package opencensus

import (
	"context"
	"fmt"
	"sync/atomic"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"

	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// Counter counts things
type Counter interface {
	Increment()
	IncrementBy(val uint64)
	Decrement()
	SetCount(count uint64)
}

type opencensusCounter struct {
	name        string
	actualCount uint64
	count       *stats.Int64Measure
	counter     *view.View
}

func (c *opencensusCounter) Increment(ctx context.Context) {
	atomic.AddUint64(&c.actualCount, 1)
	stats.Record(ctx, c.count.M(1))
}

func (c *opencensusCounter) IncrementBy(ctx context.Context, val uint64) {
	atomic.AddUint64(&c.actualCount, val)
	stats.Record(ctx, c.count.M(int64(val)))
}

func (c *opencensusCounter) Decrement(ctx context.Context) {
	atomic.AddUint64(&c.actualCount, ^uint64(0))
	stats.Record(ctx, c.count.M(-1))
}

func (c *opencensusCounter) SetCount(ctx context.Context, count uint64) {
	stats.Record(ctx, c.count.M(int64(c.actualCount)*-1))
	c.actualCount = count
	// stats.Record(ctx, c.count.M(1))
}

// ProvideUnitCounterProvider provides UnitCounter providers
func ProvideUnitCounterProvider() metrics.UnitCounterProvider {
	return ProvideUnitCounter
}

// ProvideUnitCounter provides a new counter
func ProvideUnitCounter(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
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
		return nil, errors.Wrap(err, "failed to register views")
	}

	return &opencensusCounter{
		name:    name,
		count:   count,
		counter: countView,
	}, nil

}
