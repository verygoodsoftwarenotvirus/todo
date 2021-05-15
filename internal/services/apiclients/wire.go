package apiclients

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"
)

var (
	// Providers are what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideConfig,
		ProvideAPIClientsService,
	)
)

// ProvideConfig converts an auth config to a local config.
func ProvideConfig(cfg *auth.Config) *config {
	return &config{
		minimumUsernameLength: cfg.MinimumUsernameLength,
		minimumPasswordLength: cfg.MinimumPasswordLength,
	}
}
