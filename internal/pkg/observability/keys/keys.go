package keys

const (
	// AuditLogEntryIDKey is the standard key for referring to an audit log entry ID.
	AuditLogEntryIDKey = "audit_log_entry_id"
	// AuditLogEntryEventTypeKey is the standard key for referring to an audit log event type.
	AuditLogEntryEventTypeKey = "event_type"
	// AccountSubscriptionPlanIDKey is the standard key for referring to an account subscription plan ID.
	AccountSubscriptionPlanIDKey = "account_subscription_plan_id"
	// PermissionsKey is the standard key for referring to an account user membership ID.
	PermissionsKey = "permissions"
	// RequesterKey is the standard key for referring to a requesting user's ID.
	RequesterKey = "requested_by"
	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account_id"
	// UserIDKey is the standard key for referring to a user ID.
	UserIDKey = "user_id"
	// UserIsAdminKey is the standard key for referring to a user's admin status.
	UserIsAdminKey = "is_admin"
	// UsernameKey is the standard key for referring to a username.
	UsernameKey = "username"
	// NameKey is the standard key for referring to a name.
	NameKey = "name"
	// FilterCreatedAfterKey is the standard key for referring to a types.QueryFilter's CreatedAfter field.
	FilterCreatedAfterKey = "filter_created_after"
	// FilterCreatedBeforeKey is the standard key for referring to a types.QueryFilter's CreatedBefore field.
	FilterCreatedBeforeKey = "filter_created_before"
	// FilterUpdatedAfterKey is the standard key for referring to a types.QueryFilter's UpdatedAfter field.
	FilterUpdatedAfterKey = "filter_updated_after"
	// FilterUpdatedBeforeKey is the standard key for referring to a types.QueryFilter's UpdatedAfter field.
	FilterUpdatedBeforeKey = "filter_updated_before"
	// FilterSortByKey is the standard key for referring to a types.QueryFilter's SortBy field.
	FilterSortByKey = "filter_sort_by"
	// FilterPageKey is the standard key for referring to a types.QueryFilter's page.
	FilterPageKey = "filter_page"
	// FilterLimitKey is the standard key for referring to a types.QueryFilter's limit.
	FilterLimitKey = "filter_limit"
	// FilterIsNilKey is the standard key for referring to a types.QueryFilter's null status.
	FilterIsNilKey = "filter_is_nil"
	// APIClientClientIDKey is the standard key for referring to an API client's database ID.
	APIClientClientIDKey = "api_client_id"
	// APIClientDatabaseIDKey is the standard key for referring to an API client's database ID.
	APIClientDatabaseIDKey = "api_client_row_id"
	// WebhookIDKey is the standard key for referring to a webhook's ID.
	WebhookIDKey = "webhook_id"
	// URLKey is the standard key for referring to a url.
	URLKey = "url"
	// RequestURIKey is the standard key for referring to an http.Request's URI.
	RequestURIKey = "request_uri"
	// ResponseStatusKey is the standard key for referring to an http.Request's URI.
	ResponseStatusKey = "response_status"
	// ReasonKey is the standard key for referring to a reason.
	ReasonKey = "reason"
	// QueryKey is the standard key for referring to a query.
	QueryKey = "query"
	// ConnectionDetailsKey is the standard key for referring to a database's URI.
	ConnectionDetailsKey = "connection_details"
	// SearchQueryKey is the standard key for referring to a search query parameter value.
	SearchQueryKey = "search_query"
	// UserAgentOSKey is the standard key for referring to a search query parameter value.
	UserAgentOSKey = "os"
	// UserAgentBotKey is the standard key for referring to a search query parameter value.
	UserAgentBotKey = "is_bot"
	// UserAgentMobileKey is the standard key for referring to a search query parameter value.
	UserAgentMobileKey = "is_mobile"
	// RollbackErrorKey is the standard key for referring to an error rolling back a transaction.
	RollbackErrorKey = "ROLLBACK_ERROR"
	// QueryErrorKey is the standard key for referring to an error building a query.
	QueryErrorKey = "QUERY_ERROR"
	// RowIDErrorKey is the standard key for referring to an error fetching a row ID.
	RowIDErrorKey = "ROW_ID_ERROR"

	// ItemIDKey is the standard key for referring to an item ID.
	ItemIDKey = "item_id"
)
