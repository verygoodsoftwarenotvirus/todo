package users

import (
	"github.com/google/wire"
)

var (
	// Providers is what we provide for dependency injectors
	Providers = wire.NewSet(
		ProvideUsersService,
	)
)
