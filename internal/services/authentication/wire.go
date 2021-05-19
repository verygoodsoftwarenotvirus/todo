package authentication

import (
	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideService,
	wire.FieldsOf(new(*Config),
		"Cookies",
		"PASETO",
	),
)
