package config

import (
	"github.com/google/wire"
)

var (
	// Providers represents this package's offering to the dependency manager.
	Providers = wire.NewSet(
		ProvideDatabaseConnection,
		ProvideSessionManager,
	)
)
