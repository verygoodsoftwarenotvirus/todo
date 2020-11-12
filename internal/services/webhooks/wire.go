package webhooks

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services.
	Providers = wire.NewSet(
		ProvideWebhooksService,
		ProvideWebhookDataServer,
	)
)

// ProvideWebhookDataServer is an arbitrary function for dependency injection's sake.
func ProvideWebhookDataServer(s *Service) models.WebhookDataServer {
	return s
}
