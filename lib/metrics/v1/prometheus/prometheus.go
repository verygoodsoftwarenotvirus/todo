package prometheus

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"

	"github.com/google/wire"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies
	Providers = wire.NewSet(
		ProvidePrometheus,
	)
)

// ProvidePrometheus provides a prometheus exporter
func ProvidePrometheus(ns metrics.Namespace) (view.Exporter, error) {
	pe, err := prometheus.NewExporter(
		prometheus.Options{
			Namespace: string(ns),
		},
	)
	if err != nil {
		return nil, err
	}

	// Ensure that we register it as a stats exporter.
	view.RegisterExporter(pe)

	return pe, nil
}
