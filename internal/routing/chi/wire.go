package chi

import "github.com/google/wire"

var (
	// Providers is what we provide to the dependency injector.
	Providers = wire.NewSet(
		NewRouter,
		NewRouteParamManager,
	)
)
