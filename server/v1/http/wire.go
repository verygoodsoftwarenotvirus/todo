package httpserver

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

var (
	// Providers is our wire superset of providers this package offers
	Providers = wire.NewSet(
		paramFetcherProviders,
		ProvideServer,
		ProvideNamespace,
		ProvideNewsmanTypeNameManipulationFunc,
	)
)

// ProvideNamespace provides a namespace
func ProvideNamespace() metrics.Namespace {
	return serverNamespace
}

// ProvideNewsmanTypeNameManipulationFunc provides an WebhookIDFetcher
func ProvideNewsmanTypeNameManipulationFunc(logger logging.Logger) newsman.TypeNameManipulationFunc {
	return func(s string) string {
		logger.WithName("events").WithValue("type_name", s).Info("event occurred")
		return s
	}
}
