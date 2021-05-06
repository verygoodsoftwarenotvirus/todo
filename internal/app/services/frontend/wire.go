package frontend

import (
	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	// Providers is what we offer to dependency injection.
	Providers = wire.NewSet(
		ProvideService,
		ProvideAuthService,
		ProvideUsersService,
	)
)

// ProvideAuthService does what I hope wire one day figures out how to do.
func ProvideAuthService(x types.AuthService) AuthService {
	return x
}

// ProvideUsersService does what I hope wire one day figures out how to do.
func ProvideUsersService(x types.UserDataService) UsersService {
	return x
}
