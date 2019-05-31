package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideItemsService,
		ProvideItemDataManager,
	)
)

// ProvideItemDataManager turns a database into an ItemDataManager
func ProvideItemDataManager(db database.Database) models.ItemDataManager {
	return db
}
