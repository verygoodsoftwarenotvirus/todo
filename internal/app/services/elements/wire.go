package elements

import "github.com/google/wire"

var (
	// Providers is what we offer to dependency injection.
	Providers = wire.NewSet(
		ProvideService,
	)
)
