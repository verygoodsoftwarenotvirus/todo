package tracing

import (
	"strconv"

	useragent "github.com/mssola/user_agent"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
)

func attachUint64ToSpan(span trace.Span, attachmentKey string, id uint64) {
	if span != nil {
		span.SetAttributes(label.String(attachmentKey, strconv.FormatUint(id, 10)))
	}
}

func attachStringToSpan(span trace.Span, key, str string) {
	if span != nil {
		span.SetAttributes(label.String(key, str))
	}
}

func attachBooleanToSpan(span trace.Span, key string, b bool) {
	if span != nil {
		span.SetAttributes(label.Bool(key, b))
	}
}

func attachToSpan(span trace.Span, key string, val interface{}) {
	switch x := val.(type) {
	case uint8:
		attachUint64ToSpan(span, key, uint64(x))
	case uint64:
		attachUint64ToSpan(span, key, x)
	case bool:
		attachBooleanToSpan(span, key, x)
	case string:
		attachStringToSpan(span, key, x)
	default:
		panic("invalid type to attach to span")
	}
}

// AttachFilterToSpan provides a consistent way to attach a filter's info to a span.
func AttachFilterToSpan(span trace.Span, page uint64, limit uint8) {
	attachToSpan(span, keys.FilterPageKey, page)
	attachToSpan(span, keys.FilterLimitKey, limit)
}

// AttachAuditLogEntryIDToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryIDToSpan(span trace.Span, entryID uint64) {
	attachToSpan(span, keys.AuditLogEntryIDKey, entryID)
}

// AttachAuditLogEntryEventTypeToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryEventTypeToSpan(span trace.Span, eventType string) {
	attachToSpan(span, keys.AuditLogEntryEventTypeKey, eventType)
}

// AttachAccountIDToSpan provides a consistent way to attach an account's ID to a span.
func AttachAccountIDToSpan(span trace.Span, accountID uint64) {
	attachToSpan(span, keys.AccountIDKey, accountID)
}

// AttachUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachUserIDToSpan(span trace.Span, userID uint64) {
	attachToSpan(span, keys.UserIDKey, userID)
}

// AttachRequestingUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachRequestingUserIDToSpan(span trace.Span, userID uint64) {
	attachToSpan(span, keys.RequesterKey, userID)
}

// AttachPlanIDToSpan provides a consistent way to attach a plan's ID to a span.
func AttachPlanIDToSpan(span trace.Span, planID uint64) {
	attachToSpan(span, keys.AccountSubscriptionPlanIDKey, planID)
}

// AttachRequestContextToSpan provides a consistent way to attach a RequestContext object to a span.
func AttachRequestContextToSpan(span trace.Span, sessionInfo *types.RequestContext) {
	if sessionInfo != nil {
		attachToSpan(span, keys.UserIDKey, sessionInfo.User.ID)
		attachToSpan(span, keys.UserIsAdminKey, sessionInfo.User.ServiceAdminPermissions.IsServiceAdmin())
	}
}

// AttachAPIClientDatabaseIDToSpan is a consistent way to attach an oauth2 client's ID to a span.
func AttachAPIClientDatabaseIDToSpan(span trace.Span, clientID uint64) {
	attachToSpan(span, keys.APIClientDatabaseIDKey, clientID)
}

// AttachAPIClientClientIDToSpan is a consistent way to attach an oauth2 client's ID to a span.
func AttachAPIClientClientIDToSpan(span trace.Span, clientID string) {
	attachToSpan(span, keys.APIClientClientIDKey, clientID)
}

// AttachUsernameToSpan provides a consistent way to attach a user's username to a span.
func AttachUsernameToSpan(span trace.Span, username string) {
	attachToSpan(span, keys.UsernameKey, username)
}

// AttachWebhookIDToSpan provides a consistent way to attach a webhook's ID to a span.
func AttachWebhookIDToSpan(span trace.Span, webhookID uint64) {
	attachToSpan(span, keys.WebhookIDKey, webhookID)
}

// AttachRequestURIToSpan attaches a given URI to a span.
func AttachRequestURIToSpan(span trace.Span, uri string) {
	attachToSpan(span, keys.RequestURIKey, uri)
}

// AttachSearchQueryToSpan attaches a given search query to a span.
func AttachSearchQueryToSpan(span trace.Span, query string) {
	attachToSpan(span, keys.SearchQueryKey, query)
}

// AttachUserAgentDataToSpan attaches a given search query to a span.
func AttachUserAgentDataToSpan(span trace.Span, ua *useragent.UserAgent) {
	if ua != nil {
		attachToSpan(span, keys.UserAgentOSKey, ua.OS())
		attachToSpan(span, keys.UserAgentMobileKey, ua.Mobile())
		attachToSpan(span, keys.UserAgentBotKey, ua.Bot())
	}
}

// AttachItemIDToSpan attaches an item ID to a given span.
func AttachItemIDToSpan(span trace.Span, itemID uint64) {
	attachToSpan(span, keys.ItemIDKey, itemID)
}
