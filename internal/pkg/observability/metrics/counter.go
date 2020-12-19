package metrics

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"sync/atomic"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/otel/metric"
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

// opentelemetryCounter is a Counter that interfaces with opentelemetry.
type opentelemetryCounter struct {
	name        string
	actualCount uint64
	meter       metric.Meter
	measure     metric.Int64SumObserver
}

func (c *opentelemetryCounter) subtractFromCount(ctx context.Context, value uint64) {
	atomic.AddUint64(&c.actualCount, ^value+1)

	c.measure.Observation(int64(-value))
}

func (c *opentelemetryCounter) addToCount(ctx context.Context, value uint64) {
	atomic.AddUint64(&c.actualCount, value)

	c.measure.Observation(int64(value))
}

// Decrement satisfies our Counter interface.
func (c *opentelemetryCounter) Decrement(ctx context.Context) {
	c.subtractFromCount(ctx, 1)
}

// Increment satisfies our Counter interface.
func (c *opentelemetryCounter) Increment(ctx context.Context) {

}

// IncrementBy satisfies our Counter interface.
func (c *opentelemetryCounter) IncrementBy(ctx context.Context, value uint64) {
	c.addToCount(ctx, value)
}

// ProvideOTelUnitCounter provides a new counter.
func ProvideOTelUnitCounter(counterName CounterName, description string) (UnitCounter, error) {
	name := fmt.Sprintf("%s_count", string(counterName))
	// Counts/groups the lengths of lines read in.

	if _, err := otel.Meter("ex.com/basic").NewInt64SumObserver(
		"runtime.go.cgo.calls",
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(goruntime.NumCgoCall())
		},
		metric.WithDescription("Number of cgo calls made by the current process"),
	); err != nil {
		return err
	}

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
