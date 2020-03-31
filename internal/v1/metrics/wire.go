package metrics

import (
	"github.com/google/wire"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies
	Providers = wire.NewSet(
		ProvideUnitCounter,
		ProvideUnitCounterProvider,
	)
)
