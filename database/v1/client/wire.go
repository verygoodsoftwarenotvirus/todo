package dbclient

import "github.com/google/wire"

var (
	// Providers represents what we provide to dependency injectors
	Providers = wire.NewSet(
		ProvideDatabaseClient,
	)
)
