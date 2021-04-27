package users

import (
	"github.com/google/wire"
)

// Providers is what we provide for dependency injectors.
var Providers = wire.NewSet(
	ProvideUsersService,
)
