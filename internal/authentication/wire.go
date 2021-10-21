package authentication

import "github.com/google/wire"

var (
	// Providers are what we offer to dependency injection.
	Providers = wire.NewSet(
		ProvideArgon2Authenticator,
	)
)
