package database

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
)

var (
	// Providers represents what we provide to dependency injectors.
	Providers = wire.NewSet(
		ProvideAuditLogEntryDataManager,
		ProvideItemDataManager,
		ProvideUserDataManager,
		ProvideAdminUserDataManager,
		ProvideOAuth2ClientDataManager,
		ProvideWebhookDataManager,
	)
)

// ProvideAuditLogEntryDataManager is an arbitrary function for dependency injection's sake.
func ProvideAuditLogEntryDataManager(db DataManager) models.AuditLogEntryDataManager {
	return db
}

// ProvideItemDataManager is an arbitrary function for dependency injection's sake.
func ProvideItemDataManager(db DataManager) models.ItemDataManager {
	return db
}

// ProvideUserDataManager is an arbitrary function for dependency injection's sake.
func ProvideUserDataManager(db DataManager) models.UserDataManager {
	return db
}

// ProvideAdminUserDataManager is an arbitrary function for dependency injection's sake.
func ProvideAdminUserDataManager(db DataManager) models.AdminUserDataManager {
	return db
}

// ProvideOAuth2ClientDataManager is an arbitrary function for dependency injection's sake.
func ProvideOAuth2ClientDataManager(db DataManager) models.OAuth2ClientDataManager {
	return db
}

// ProvideWebhookDataManager is an arbitrary function for dependency injection's sake.
func ProvideWebhookDataManager(db DataManager) models.WebhookDataManager {
	return db
}
