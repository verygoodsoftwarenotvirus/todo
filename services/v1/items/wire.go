package items

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services.
	Providers = wire.NewSet(
		ProvideItemsService,
		ProvideItemDataServer,
		ProvideItemsServiceSearchIndex,
	)
)

// ProvideItemDataServer is an arbitrary function for dependency injection's sake.
func ProvideItemDataServer(s *Service) models.ItemDataServer {
	return s
}
