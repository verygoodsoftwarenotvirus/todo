package events

import "github.com/google/wire"

var (
	// Providers is what we offer up for dependency injection.
	Providers = wire.NewSet(
		ProvidePublisher,
		ProvideSubscriber,
	)
)
