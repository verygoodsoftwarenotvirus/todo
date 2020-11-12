package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideItemsService,
	ProvideItemDataServer,
	ProvideItemsServiceSearchIndex,
)

// ProvideItemDataServer is an arbitrary function for dependency injection's sake.
func ProvideItemDataServer(s *Service) types.ItemDataServer {
	return s
}
