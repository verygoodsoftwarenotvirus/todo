package elasticsearch

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies.
	Providers = wire.NewSet(
		ProvideIndexManagerProvider,
	)
)

// ProvideIndexManagerProvider is a wrapper around NewIndexManager.
func ProvideIndexManagerProvider() search.IndexManagerProvider {
	return NewIndexManager
}
