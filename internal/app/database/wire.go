package database

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

var (
	// Providers represents what we provide to dependency injectors.
	Providers = wire.NewSet(
		ProvideAuthAuditManager,
		ProvideAuditLogEntryDataManager,
		ProvideItemDataManager,
		ProvideItemAuditManager,
		ProvideUserDataManager,
		ProvideUserAuditManager,
		ProvideAdminUserDataManager,
		ProvideOAuth2ClientDataManager,
		ProvideOAuth2ClientAuditManager,
		ProvideWebhookDataManager,
		ProvideWebhookAuditManager,
	)
)

// ProvideAuthAuditManager is an arbitrary function for dependency injection's sake.
func ProvideAuthAuditManager(db DataManager) types.AuthAuditManager {
	return db
}

// ProvideAuditLogEntryDataManager is an arbitrary function for dependency injection's sake.
func ProvideAuditLogEntryDataManager(db DataManager) types.AuditLogDataManager {
	return db
}

// ProvideItemDataManager is an arbitrary function for dependency injection's sake.
func ProvideItemDataManager(db DataManager) types.ItemDataManager {
	return db
}

// ProvideItemAuditManager is an arbitrary function for dependency injection's sake.
func ProvideItemAuditManager(db DataManager) types.ItemAuditManager {
	return db
}

// ProvideUserDataManager is an arbitrary function for dependency injection's sake.
func ProvideUserDataManager(db DataManager) types.UserDataManager {
	return db
}

// ProvideUserAuditManager is an arbitrary function for dependency injection's sake.
func ProvideUserAuditManager(db DataManager) types.UserAuditManager {
	return db
}

// ProvideAdminUserDataManager is an arbitrary function for dependency injection's sake.
func ProvideAdminUserDataManager(db DataManager) types.AdminUserDataManager {
	return db
}

// ProvideOAuth2ClientDataManager is an arbitrary function for dependency injection's sake.
func ProvideOAuth2ClientDataManager(db DataManager) types.OAuth2ClientDataManager {
	return db
}

// ProvideOAuth2ClientAuditManager is an arbitrary function for dependency injection's sake.
func ProvideOAuth2ClientAuditManager(db DataManager) types.OAuth2ClientAuditManager {
	return db
}

// ProvideWebhookDataManager is an arbitrary function for dependency injection's sake.
func ProvideWebhookDataManager(db DataManager) types.WebhookDataManager {
	return db
}

// ProvideWebhookAuditManager is an arbitrary function for dependency injection's sake.
func ProvideWebhookAuditManager(db DataManager) types.WebhookAuditManager {
	return db
}
