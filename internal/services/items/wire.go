package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideItemsService,
	ProvideItemDataServer,
	ProvideItemsServiceSearchIndex,
)

// ProvideItemDataServer is an arbitrary function for dependency injection's sake.
func ProvideItemDataServer(s *Service) models.ItemDataServer {
	return s
}
