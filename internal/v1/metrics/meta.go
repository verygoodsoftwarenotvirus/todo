package metrics

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	// MetricAggregationMeasurement keeps track of how much time we spend collecting metrics
	MetricAggregationMeasurement = stats.Int64(
		"metrics_aggregation_time",
		"cumulative time in nanoseconds spent aggregating metrics",
		stats.UnitDimensionless,
	)

	// MetricAggregationMeasurementView is the corresponding view for the above metric
	MetricAggregationMeasurementView = &view.View{
		Name:        "metrics_aggregation_time",
		Measure:     MetricAggregationMeasurement,
		Description: "cumulative time in nanoseconds spent aggregating metrics",
		Aggregation: view.LastValue(),
	}
)

// RegisterDefaultViews registers default runtime views
func RegisterDefaultViews() error {
	return view.Register(DefaultRuntimeViews...)
}
