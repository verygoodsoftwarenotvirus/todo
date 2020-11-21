package bcrypt

import "github.com/google/wire"

var (
	// Providers is what we provide to the dependency injector.
	Providers = wire.NewSet(
		ProvideHashCost,
	)
)

// ProvideHashCost provides a BcryptHashCost.
func ProvideHashCost() HashCost {
	return DefaultHashCost
}
