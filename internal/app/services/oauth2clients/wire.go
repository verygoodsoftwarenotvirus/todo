package oauth2clients

import (
	"github.com/google/wire"
)

var (
	// Providers are what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideOAuth2ClientsService,
	)
)
