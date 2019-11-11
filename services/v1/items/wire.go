package items

import (
	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideItemsService,
		ProvideItemDataManager,
		ProvideItemDataServer,
	)
)

// ProvideItemDataManager turns a database into an ItemDataManager
func ProvideItemDataManager(db database.Database) models.ItemDataManager {
	return db
}

// ProvideItemDataServer is an arbitrary function for dependency injection's sake
func ProvideItemDataServer(s *Service) models.ItemDataServer {
	return s
}
