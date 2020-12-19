package httpserver

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

// Providers is our wire superset of providers this package offers.
var Providers = wire.NewSet(
	ProvideServer,
	ProvideNamespace,
	ProvideNewsmanTypeNameManipulationFunc,
)

// ProvideNamespace provides a namespace.
func ProvideNamespace() metrics.Namespace {
	return serverNamespace
}

// ProvideNewsmanTypeNameManipulationFunc provides an WebhookIDFetcher.
func ProvideNewsmanTypeNameManipulationFunc() newsman.TypeNameManipulationFunc {
	return func(s string) string {
		return s
	}
}
