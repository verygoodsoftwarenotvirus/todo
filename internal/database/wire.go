package database

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	// Providers represents what we provide to dependency injectors.
	Providers = wire.NewSet(
		ProvideItemDataManager,
		ProvideUserDataManager,
		ProvideAdminUserDataManager,
		ProvideAccountDataManager,
		ProvideAccountUserMembershipDataManager,
		ProvideAPIClientDataManager,
		ProvideWebhookDataManager,
	)
)

// ProvideAccountDataManager is an arbitrary function for dependency injection's sake.
func ProvideAccountDataManager(db DataManager) types.AccountDataManager {
	return db
}

// ProvideAccountUserMembershipDataManager is an arbitrary function for dependency injection's sake.
func ProvideAccountUserMembershipDataManager(db DataManager) types.AccountUserMembershipDataManager {
	return db
}

// ProvideItemDataManager is an arbitrary function for dependency injection's sake.
func ProvideItemDataManager(db DataManager) types.ItemDataManager {
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

// ProvideAPIClientDataManager is an arbitrary function for dependency injection's sake.
func ProvideAPIClientDataManager(db DataManager) types.APIClientDataManager {
	return db
}

// ProvideWebhookDataManager is an arbitrary function for dependency injection's sake.
func ProvideWebhookDataManager(db DataManager) types.WebhookDataManager {
	return db
}
