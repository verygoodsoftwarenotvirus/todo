package webhooks

import (
	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideWebhooksService,
		ProvideWebhookDataManager,
		ProvideWebhookDataServer,
	)
)

// ProvideWebhookDataManager is an arbitrary function for dependency injection's sake
func ProvideWebhookDataManager(db database.Database) models.WebhookDataManager {
	return db
}

// ProvideWebhookDataServer is an arbitrary function for dependency injection's sake
func ProvideWebhookDataServer(s *Service) models.WebhookDataServer {
	return s
}
