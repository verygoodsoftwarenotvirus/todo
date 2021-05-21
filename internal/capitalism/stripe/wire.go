package stripe

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
)

var (
	// Providers represents this package's offering to the dependency manager.
	Providers = wire.NewSet(
		ProvideAPIKey,
		ProvideWebhookSecret,
		NewStripePaymentManager,
	)
)

// ProvideAPIKey is an arbitrary wrapper for wire.
func ProvideAPIKey(cfg *capitalism.StripeConfig) APIKey {
	return APIKey(cfg.APIKey)
}

// ProvideWebhookSecret is an arbitrary wrapper for wire.
func ProvideWebhookSecret(cfg *capitalism.StripeConfig) WebhookSecret {
	return WebhookSecret(cfg.WebhookSecret)
}
