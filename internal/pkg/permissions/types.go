package permissions

// ServiceAdminPermissionChecker returns whether or not a given permission applies to a user.
type ServiceAdminPermissionChecker interface {
	IsServiceAdmin() bool
	CanCycleCookieSecrets() bool
	CanBanUsers() bool
	CanTerminateAccounts() bool
}

// ServiceAdminPermissionsSummary summarizes a user's permissions.
type ServiceAdminPermissionsSummary struct {
	CanCycleCookieSecrets bool `json:"canCycleCookieSecret"`
	CanBanUsers           bool `json:"canBanUsers"`
	CanTerminateAccounts  bool `json:"canTerminateAccounts"`
}

// ServiceUserPermissionChecker returns whether or not a given permission applies to a user.
type ServiceUserPermissionChecker interface {
	// API Clients
	CanReadAPIClients() bool
	CanCreateAPIClients() bool
	CanUpdateAPIClients() bool
	CanArchiveAPIClients() bool

	// Webhooks
	CanReadWebhooks() bool
	CanCreateWebhooks() bool
	CanUpdateWebhooks() bool
	CanArchiveWebhooks() bool

	// Items
	CanReadItems() bool
	CanCreateItems() bool
	CanUpdateItems() bool
	CanArchiveItems() bool
}
