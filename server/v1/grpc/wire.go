package grpcserver

import (
	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideGRPCServer,
	)
)
