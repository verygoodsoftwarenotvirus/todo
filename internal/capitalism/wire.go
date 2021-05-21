package capitalism

import (
	"github.com/google/wire"
)

var (
	// Providers are what we provide to dependency injection.
	Providers = wire.NewSet(
		wire.FieldsOf(new(*Config),
			"Stripe",
		),
	)
)
