package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideItemsService,
		ProvideItemDataManager,
		ProvideReporter,
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

// ProvideReporter is an obligatory function that hopefully wire will eliminate for me one day
func ProvideReporter(n *newsman.Newsman) newsman.Reporter {
	return n
}
