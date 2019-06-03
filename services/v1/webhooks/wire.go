package webhooks

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideWebhooksService,
		ProvideWebhookDataManager,
	)
)

// ProvideWebhookDataManager turns a database into an ItemDataManager
func ProvideWebhookDataManager(db database.Database) models.WebhookDataManager {
	return db
}
