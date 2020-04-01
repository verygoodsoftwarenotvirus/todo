package httpserver

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"

	"github.com/google/wire"
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
func ProvideNewsmanTypeNameManipulationFunc() newsman.TypeNameManipulationFunc {
	return func(s string) string {
		return s
	}
}
