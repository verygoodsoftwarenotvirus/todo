package database

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

var (
	// Providers represents what we provide to dependency injectors.
	Providers = wire.NewSet(
		ProvideAdminAuditManager,
		ProvideAuthAuditManager,
		ProvideAuditLogEntryDataManager,
		ProvidePlanDataManager,
		ProvidePlanAuditManager,
		ProvideItemDataManager,
		ProvideItemAuditManager,
		ProvideUserDataManager,
		ProvideAdminUserDataManager,
		ProvideAccountDataManager,
		ProvideAccountAuditManager,
		ProvideDelegatedClientDataManager,
		ProvideDelegatedClientAuditManager,
		ProvideOAuth2ClientDataManager,
		ProvideOAuth2ClientAuditManager,
		ProvideWebhookDataManager,
		ProvideWebhookAuditManager,
	)
)

// ProvideAdminAuditManager is an arbitrary function for dependency injection's sake.
func ProvideAdminAuditManager(db DataManager) types.AdminAuditManager {
	return db
}

// ProvideAuthAuditManager is an arbitrary function for dependency injection's sake.
func ProvideAuthAuditManager(db DataManager) types.AuthAuditManager {
	return db
}

// ProvideAuditLogEntryDataManager is an arbitrary function for dependency injection's sake.
func ProvideAuditLogEntryDataManager(db DataManager) types.AuditLogEntryDataManager {
	return db
}

// ProvidePlanDataManager is an arbitrary function for dependency injection's sake.
func ProvidePlanDataManager(db DataManager) types.AccountSubscriptionPlanDataManager {
	return db
}

// ProvidePlanAuditManager is an arbitrary function for dependency injection's sake.
func ProvidePlanAuditManager(db DataManager) types.AccountSubscriptionPlanAuditManager {
	return db
}

// ProvideAccountDataManager is an arbitrary function for dependency injection's sake.
func ProvideAccountDataManager(db DataManager) types.AccountDataManager {
	return db
}

// ProvideAccountAuditManager is an arbitrary function for dependency injection's sake.
func ProvideAccountAuditManager(db DataManager) types.AccountAuditManager {
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

// ProvideAdminUserDataManager is an arbitrary function for dependency injection's sake.
func ProvideAdminUserDataManager(db DataManager) types.AdminUserDataManager {
	return db
}

// ProvideDelegatedClientDataManager is an arbitrary function for dependency injection's sake.
func ProvideDelegatedClientDataManager(db DataManager) types.DelegatedClientDataManager {
	return db
}

// ProvideDelegatedClientAuditManager is an arbitrary function for dependency injection's sake.
func ProvideDelegatedClientAuditManager(db DataManager) types.DelegatedClientAuditManager {
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
