package keys

const (
	// AuditLogEntryIDKey is the standard key for referring to an audit log entry ID in a log.
	AuditLogEntryIDKey = "audit_log_entry_id"
	// AuditLogEntryEventTypeKey is the standard key for referring to an audit log event type in a log.
	AuditLogEntryEventTypeKey = "event_type"
	// AccountSubscriptionPlanIDKey is the standard key for referring to a plan ID in a log.
	AccountSubscriptionPlanIDKey = "plan_id"
	// AccountIDKey is the standard key for referring to an account ID in a log.
	AccountIDKey = "account_id"
	// UserIDKey is the standard key for referring to a user ID in a log.
	UserIDKey = "user_id"
	// UserIsAdminKey is the standard key for referring to a user's admin status in a log.
	UserIsAdminKey = "is_admin"
	// UsernameKey is the standard key for referring to a username in a log.
	UsernameKey = "username"
	// FilterPageKey is the standard key for referring to a types.QueryFilter's page in a log.
	FilterPageKey = "filter_page"
	// FilterLimitKey is the standard key for referring to a types.QueryFilter's limit in a log.
	FilterLimitKey = "filter_limit"
	// FilterIsNilKey is the standard key for referring to a types.QueryFilter's null status in a log.
	FilterIsNilKey = "filter_is_nil"
	// DelegatedClientIDKey is the standard key for referring to a delegated client's database ID in a log.
	DelegatedClientIDKey = "delegated_client_id"
	// OAuth2ClientDatabaseIDKey is the standard key for referring to an oauth2 client's database ID in a log.
	OAuth2ClientDatabaseIDKey = "oauth2_client_database_id"
	// OAuth2ClientIDKey is the standard key for referring to an oauth2 client's ClientID in a log.
	OAuth2ClientIDKey = "client_id"
	// WebhookIDKey is the standard key for referring to a webhook's ID in a log.
	WebhookIDKey = "webhook_id"
	// URLKey is the standard key for referring to a url in a log.
	URLKey = "url"
	// RequestURIKey is the standard key for referring to an http.Request's URI in a log.
	RequestURIKey = "request_uri"
	// ResponseStatusKey is the standard key for referring to an http.Request's URI in a log.
	ResponseStatusKey = "response_status"
	// QueryKey is the standard key for referring to a query in a log.
	QueryKey = "query"
	// ConnectionDetailsKey is the standard key for referring to a database's URI in a log.
	ConnectionDetailsKey = "connection_details"
	// SearchQueryKey is the standard key for referring to a search query parameter value.
	SearchQueryKey = "search_query"
	// RollbackErrorKey is the standard key for referring to an error rolling back a transaction.
	RollbackErrorKey = "ROLLBACK_ERROR"
	// QueryErrorKey is the standard key for referring to an error building a query.
	QueryErrorKey = "QUERY_ERROR"
	// RowIDErrorKey is the standard key for referring to an error fetching a row ID.
	RowIDErrorKey = "ROW_ID_ERROR"

	// ItemIDKey is the standard key for referring to an item ID in a log.
	ItemIDKey = "item_id"
)
