package bleve

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	"github.com/google/wire"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies.
	Providers = wire.NewSet(
		ProvideBleveIndexManagerProvider,
	)
)

// ProvideBleveIndexManagerProvider is a wrapper around NewBleveIndexManager.
func ProvideBleveIndexManagerProvider() search.IndexManagerProvider {
	return NewBleveIndexManager
}
