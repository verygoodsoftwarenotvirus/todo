package tracing

import (
	"net/url"

	useragent "github.com/mssola/user_agent"
	"go.opentelemetry.io/otel/codes"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func attachUint8ToSpan(span trace.Span, attachmentKey string, id uint8) {
	if span != nil {
		span.SetAttributes(attribute.Int64(attachmentKey, int64(id)))
	}
}

func attachUint64ToSpan(span trace.Span, attachmentKey string, id uint64) {
	if span != nil {
		span.SetAttributes(attribute.Int64(attachmentKey, int64(id)))
	}
}

func attachStringToSpan(span trace.Span, key, str string) {
	if span != nil {
		span.SetAttributes(attribute.String(key, str))
	}
}

func attachBooleanToSpan(span trace.Span, key string, b bool) {
	if span != nil {
		span.SetAttributes(attribute.Bool(key, b))
	}
}

func attachSliceToSpan(span trace.Span, key string, slice interface{}) {
	span.SetAttributes(attribute.Array(key, slice))
}

func attachToSpan(span trace.Span, key string, val interface{}) {
	switch x := val.(type) {
	case uint8:
		attachUint8ToSpan(span, key, x)
	case uint64:
		attachUint64ToSpan(span, key, x)
	case bool:
		attachBooleanToSpan(span, key, x)
	case string:
		attachStringToSpan(span, key, x)
	case error:
		attachStringToSpan(span, key, x.Error())
	default:
		panic("invalid type to attach to span")
	}
}

// AttachToSpan allows a user to attach any value to a span.
func AttachToSpan(span trace.Span, key string, val interface{}) {
	if span != nil {
		span.SetAttributes(attribute.Any(key, val))
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

// AttachAccountSubscriptionPlanIDToSpan provides a consistent way to attach a plan's ID to a span.
func AttachAccountSubscriptionPlanIDToSpan(span trace.Span, planID uint64) {
	attachToSpan(span, keys.AccountSubscriptionPlanIDKey, planID)
}

// AttachRequestContextToSpan provides a consistent way to attach a RequestContext object to a span.
func AttachRequestContextToSpan(span trace.Span, sessionInfo *types.RequestContext) {
	if sessionInfo != nil {
		attachToSpan(span, keys.UserIDKey, sessionInfo.User.ID)
		attachToSpan(span, keys.UserIsAdminKey, sessionInfo.User.ServiceAdminPermissions.IsServiceAdmin())
	}
}

// AttachAPIClientDatabaseIDToSpan is a consistent way to attach an API client's database row ID to a span.
func AttachAPIClientDatabaseIDToSpan(span trace.Span, clientID uint64) {
	attachToSpan(span, keys.APIClientDatabaseIDKey, clientID)
}

// AttachAPIClientClientIDToSpan is a consistent way to attach an API client's ID to a span.
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

// AttachURLToSpan attaches a given URI to a span.
func AttachURLToSpan(span trace.Span, u *url.URL) {
	attachToSpan(span, keys.RequestURIKey, u.String())
}

// AttachRequestURIToSpan attaches a given URI to a span.
func AttachRequestURIToSpan(span trace.Span, uri string) {
	attachToSpan(span, keys.RequestURIKey, uri)
}

// AttachErrorToSpan attaches a given error to a span.
func AttachErrorToSpan(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
}

// AttachDatabaseQueryToSpan attaches a given search query to a span.
func AttachDatabaseQueryToSpan(span trace.Span, query, queryDescription string, args []interface{}) {
	attachToSpan(span, keys.QueryKey, query)
	attachToSpan(span, "query_description", queryDescription)

	if args != nil {
		attachSliceToSpan(span, "query_args", args)
	}
}

// AttachQueryFilterToSpan attaches a given query filter to a span.
func AttachQueryFilterToSpan(span trace.Span, filter *types.QueryFilter) {
	if filter != nil {
		attachUint8ToSpan(span, keys.FilterLimitKey, filter.Limit)
		attachUint64ToSpan(span, keys.FilterPageKey, filter.Page)
		attachUint64ToSpan(span, keys.FilterCreatedAfterKey, filter.CreatedAfter)
		attachUint64ToSpan(span, keys.FilterCreatedBeforeKey, filter.CreatedBefore)
		attachUint64ToSpan(span, keys.FilterUpdatedAfterKey, filter.UpdatedAfter)
		attachUint64ToSpan(span, keys.FilterUpdatedBeforeKey, filter.UpdatedBefore)
		attachStringToSpan(span, keys.FilterSortByKey, string(filter.SortBy))
	} else {
		attachBooleanToSpan(span, keys.FilterIsNilKey, true)
	}
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
