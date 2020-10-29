package users

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is what we provide for dependency injectors.
	Providers = wire.NewSet(
		ProvideUsersService,
		ProvideUserDataServer,
	)
)

// ProvideUserDataServer is an arbitrary function for dependency injection's sake.
func ProvideUserDataServer(s *Service) models.UserDataServer {
	return s
}
