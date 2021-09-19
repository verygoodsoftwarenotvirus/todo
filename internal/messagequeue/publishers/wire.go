package publishers

import (
	"github.com/google/wire"
)

var (
	// Providers are what we provide to dependency injection.
	Providers = wire.NewSet(
		ProvidePublisherProvider,
		wire.FieldsOf(
			new(Config),
			"Provider",
			"QueueAddress",
		),
	)
)
