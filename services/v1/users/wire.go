package users

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is what we provide for dependency injectors
	Providers = wire.NewSet(
		ProvideUsersService,
		ProvideUserDataManager,
	)
)

// ProvideUserDataManager is an arbitrary function for dependency injection's sake
func ProvideUserDataManager(db database.Database) models.UserDataManager {
	return db
}