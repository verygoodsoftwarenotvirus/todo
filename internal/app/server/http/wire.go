package httpserver

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"

	"github.com/google/wire"
)

// Providers is our wire superset of providers this package offers.
var Providers = wire.NewSet(
	ProvideServer,
	ProvideNamespace,
)

// ProvideNamespace provides a namespace.
func ProvideNamespace() metrics.Namespace {
	return serverNamespace
}
