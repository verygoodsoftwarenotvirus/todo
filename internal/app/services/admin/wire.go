package admin

import (
	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideService,
)
