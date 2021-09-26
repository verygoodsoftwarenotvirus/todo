package bleve

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
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
