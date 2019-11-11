package config

import "github.com/google/wire"

// BEGIN it'd be neat if wire could do this for me one day.

// ProvideConfigServerSettings is an obligatory function that
// we're required to have because wire doesn't do it for us.
func ProvideConfigServerSettings(c *ServerConfig) ServerSettings {
	return c.Server
}

// ProvideConfigAuthSettings is an obligatory function that
// we're required to have because wire doesn't do it for us.
func ProvideConfigAuthSettings(c *ServerConfig) AuthSettings {
	return c.Auth
}

// ProvideConfigDatabaseSettings is an obligatory function that
// we're required to have because wire doesn't do it for us.
func ProvideConfigDatabaseSettings(c *ServerConfig) DatabaseSettings {
	return c.Database
}

// ProvideConfigFrontendSettings is an obligatory function that
// we're required to have because wire doesn't do it for us.
func ProvideConfigFrontendSettings(c *ServerConfig) FrontendSettings {
	return c.Frontend
}

// END it'd be neat if wire could do this for me one day.

var (
	// Providers represents this package's offering to the dependency manager
	Providers = wire.NewSet(
		ProvideConfigServerSettings,
		ProvideConfigAuthSettings,
		ProvideConfigDatabaseSettings,
		ProvideConfigFrontendSettings,
	)
)
