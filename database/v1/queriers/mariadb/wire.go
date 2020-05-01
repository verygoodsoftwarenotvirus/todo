package mariadb

import (
	"github.com/google/wire"
)

var (
	// Providers is what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideMariaDBConnection,
		ProvideMariaDB,
	)
)
