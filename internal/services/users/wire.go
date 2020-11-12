package users

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/google/wire"
)

// Providers is what we provide for dependency injectors.
var Providers = wire.NewSet(
	ProvideUsersService,
	ProvideUserDataServer,
)

// ProvideUserDataServer is an arbitrary function for dependency injection's sake.
func ProvideUserDataServer(s *Service) models.UserDataServer {
	return s
}
