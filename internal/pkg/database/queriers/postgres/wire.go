package postgres

import (
	"github.com/google/wire"
)

// Providers is what we provide for dependency injection.
var Providers = wire.NewSet(
	ProvidePostgresDB,
	ProvidePostgres,
)
