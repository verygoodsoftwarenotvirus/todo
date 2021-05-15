package server

import (
	"github.com/google/wire"
)

// Providers is our wire superset of providers this package offers.
var Providers = wire.NewSet(
	ProvideHTTPServer,
)
