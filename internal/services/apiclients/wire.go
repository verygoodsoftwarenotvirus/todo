package apiclients

import (
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"github.com/google/wire"
)

var (
	// Providers are what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideConfig,
		ProvideAPIClientsService,
	)
)

// ProvideConfig converts an auth config to a local config.
func ProvideConfig(cfg *authservice.Config) *config {
	return &config{
		minimumUsernameLength: cfg.MinimumUsernameLength,
		minimumPasswordLength: cfg.MinimumPasswordLength,
	}
}
