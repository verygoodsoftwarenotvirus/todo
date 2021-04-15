package tracing

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	useragent "github.com/mssola/user_agent"

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

// AttachToSpan allows a user to attach any value to a span.
func AttachToSpan(span trace.Span, key string, val interface{}) {
	if span != nil {
		span.SetAttributes(attribute.Any(key, val))
	}
}

// AttachFilterToSpan provides a consistent way to attach a filter's info to a span.
func AttachFilterToSpan(span trace.Span, page uint64, limit uint8, sortBy string) {
	attachUint64ToSpan(span, keys.FilterPageKey, page)
	attachUint8ToSpan(span, keys.FilterLimitKey, limit)
	attachStringToSpan(span, keys.FilterSortByKey, sortBy)
}

// AttachAuditLogEntryIDToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryIDToSpan(span trace.Span, entryID uint64) {
	attachUint64ToSpan(span, keys.AuditLogEntryIDKey, entryID)
}

// AttachAuditLogEntryEventTypeToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryEventTypeToSpan(span trace.Span, eventType string) {
	attachStringToSpan(span, keys.AuditLogEntryEventTypeKey, eventType)
}

// AttachAccountIDToSpan provides a consistent way to attach an account's ID to a span.
func AttachAccountIDToSpan(span trace.Span, accountID uint64) {
	attachUint64ToSpan(span, keys.AccountIDKey, accountID)
}

// AttachRequestingUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachRequestingUserIDToSpan(span trace.Span, userID uint64) {
	attachUint64ToSpan(span, keys.RequesterIDKey, userID)
}

// AttachAccountSubscriptionPlanIDToSpan provides a consistent way to attach a plan's ID to a span.
func AttachAccountSubscriptionPlanIDToSpan(span trace.Span, planID uint64) {
	attachUint64ToSpan(span, keys.AccountSubscriptionPlanIDKey, planID)
}

// AttachChangeSummarySpan provides a consistent way to attach a SessionContextData object to a span.
func AttachChangeSummarySpan(span trace.Span, typeName string, changes []*types.FieldChangeSummary) {
	for i, change := range changes {
		span.SetAttributes(attribute.Any(fmt.Sprintf("%s.field_changes.%d", typeName, i), change))
	}
}

// AttachSessionContextDataToSpan provides a consistent way to attach a SessionContextData object to a span.
func AttachSessionContextDataToSpan(span trace.Span, sessionCtxData *types.SessionContextData) {
	if sessionCtxData != nil {
		attachUint64ToSpan(span, keys.RequesterIDKey, sessionCtxData.Requester.ID)
		attachUint64ToSpan(span, keys.ActiveAccountIDKey, sessionCtxData.ActiveAccountID)
		attachBooleanToSpan(span, keys.UserIsAdminKey, sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin())
	}
}

// AttachAPIClientDatabaseIDToSpan is a consistent way to attach an API client's database row ID to a span.
func AttachAPIClientDatabaseIDToSpan(span trace.Span, clientID uint64) {
	attachUint64ToSpan(span, keys.APIClientDatabaseIDKey, clientID)
}

// AttachAPIClientClientIDToSpan is a consistent way to attach an API client's ID to a span.
func AttachAPIClientClientIDToSpan(span trace.Span, clientID string) {
	attachStringToSpan(span, keys.APIClientClientIDKey, clientID)
}

// AttachUserToSpan provides a consistent way to attach a user to a span.
func AttachUserToSpan(span trace.Span, user *types.User) {
	if user != nil {
		AttachUserIDToSpan(span, user.ID)
		AttachUsernameToSpan(span, user.Username)
	}
}

// AttachUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachUserIDToSpan(span trace.Span, userID uint64) {
	attachUint64ToSpan(span, keys.UserIDKey, userID)
}

// AttachUsernameToSpan provides a consistent way to attach a user's username to a span.
func AttachUsernameToSpan(span trace.Span, username string) {
	attachStringToSpan(span, keys.UsernameKey, username)
}

// AttachWebhookIDToSpan provides a consistent way to attach a webhook's ID to a span.
func AttachWebhookIDToSpan(span trace.Span, webhookID uint64) {
	attachUint64ToSpan(span, keys.WebhookIDKey, webhookID)
}

// AttachURLToSpan attaches a given URI to a span.
func AttachURLToSpan(span trace.Span, u *url.URL) {
	attachStringToSpan(span, keys.RequestURIKey, u.String())
}

// AttachRequestURIToSpan attaches a given URI to a span.
func AttachRequestURIToSpan(span trace.Span, uri string) {
	attachStringToSpan(span, keys.RequestURIKey, uri)
}

// AttachRequestToSpan attaches a given *http.Request to a span.
func AttachRequestToSpan(span trace.Span, req *http.Request) {
	if req != nil {
		attachStringToSpan(span, keys.RequestURIKey, req.URL.String())
		attachStringToSpan(span, keys.RequestMethodKey, req.Method)

		for k, v := range req.Header {
			attachSliceToSpan(span, fmt.Sprintf("%s.%s", keys.RequestHeadersKey, k), v)
		}
	}
}

// AttachResponseToSpan attaches a given *http.Response to a span.
func AttachResponseToSpan(span trace.Span, res *http.Response) {
	if res != nil {
		AttachRequestToSpan(span, res.Request)

		span.SetAttributes(attribute.Int(keys.ResponseStatusKey, res.StatusCode))

		for k, v := range res.Header {
			attachSliceToSpan(span, fmt.Sprintf("%s.%s", keys.ResponseHeadersKey, k), v)
		}
	}
}

// AttachErrorToSpan attaches a given error to a span.
func AttachErrorToSpan(span trace.Span, description string, err error) {
	if err != nil {
		span.RecordError(
			err,
			trace.WithTimestamp(time.Now()),
			trace.WithAttributes(attribute.String("error.description", description)),
		)
	}
}

// AttachDatabaseQueryToSpan attaches a given search query to a span.
func AttachDatabaseQueryToSpan(span trace.Span, queryDescription, query string, args []interface{}) {
	attachStringToSpan(span, keys.DatabaseQueryKey, query)
	attachStringToSpan(span, "query_description", queryDescription)

	for i, arg := range args {
		span.SetAttributes(attribute.Any(fmt.Sprintf("query_args_%d", i), arg))
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
	attachStringToSpan(span, keys.SearchQueryKey, query)
}

// AttachUserAgentDataToSpan attaches a given search query to a span.
func AttachUserAgentDataToSpan(span trace.Span, ua *useragent.UserAgent) {
	if ua != nil {
		attachStringToSpan(span, keys.UserAgentOSKey, ua.OS())
		attachBooleanToSpan(span, keys.UserAgentMobileKey, ua.Mobile())
		attachBooleanToSpan(span, keys.UserAgentBotKey, ua.Bot())
	}
}

// AttachItemIDToSpan attaches an item ID to a given span.
func AttachItemIDToSpan(span trace.Span, itemID uint64) {
	attachUint64ToSpan(span, keys.ItemIDKey, itemID)
}
