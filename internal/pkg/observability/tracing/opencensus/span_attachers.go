package opencensus

import (
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"

	"go.opencensus.io/trace"
)

func attachUint64ToSpan(span *trace.Span, attachmentKey string, id uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute(attachmentKey, strconv.FormatUint(id, 10)))
	}
}

func attachIntToSpan(span *trace.Span, attachmentKey string, id int) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute(attachmentKey, strconv.FormatInt(int64(id), 10)))
	}
}

func attachStringToSpan(span *trace.Span, key, str string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute(key, str))
	}
}

func attachBooleanToSpan(span *trace.Span, key string, b bool) {
	if span != nil {
		span.AddAttributes(trace.BoolAttribute(key, b))
	}
}

// AttachFilterToSpan provides a consistent way to attach a filter's info to a span.
func AttachFilterToSpan(span *trace.Span, page uint64, limit uint8) {
	span.AddAttributes(
		trace.StringAttribute(keys.FilterPageKey, strconv.FormatUint(page, 10)),
		trace.StringAttribute(keys.FilterLimitKey, strconv.FormatUint(uint64(limit), 10)),
	)
}

// AttachAuditLogEntryIDToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryIDToSpan(span *trace.Span, entryID uint64) {
	attachUint64ToSpan(span, keys.AuditLogEntryIDKey, entryID)
}

// AttachAuditLogEntryEventTypeToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryEventTypeToSpan(span *trace.Span, eventType int) {
	attachIntToSpan(span, keys.AuditLogEntryEventTypeKey, eventType)
}

// AttachItemIDToSpan attaches an item ID to a given span.
func AttachItemIDToSpan(span *trace.Span, itemID uint64) {
	attachUint64ToSpan(span, keys.ItemIDKey, itemID)
}

// AttachUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachUserIDToSpan(span *trace.Span, userID uint64) {
	attachUint64ToSpan(span, keys.UserIDKey, userID)
}

// AttachSessionInfoToSpan provides a consistent way to attach a SessionInfo object to a span.
func AttachSessionInfoToSpan(span *trace.Span, userID uint64, userIsAdmin bool) {
	attachUint64ToSpan(span, keys.UserIDKey, userID)
	attachBooleanToSpan(span, keys.UserIsAdminKey, userIsAdmin)
}

// AttachOAuth2ClientDatabaseIDToSpan is a consistent way to attach an oauth2 client's ID to a span.
func AttachOAuth2ClientDatabaseIDToSpan(span *trace.Span, oauth2ClientID uint64) {
	attachUint64ToSpan(span, keys.OAuth2ClientDatabaseIDKey, oauth2ClientID)
}

// AttachOAuth2ClientIDToSpan is a consistent way to attach an oauth2 client's Client ID to a span.
func AttachOAuth2ClientIDToSpan(span *trace.Span, clientID string) {
	attachStringToSpan(span, keys.OAuth2ClientIDKey, clientID)
}

// AttachUsernameToSpan provides a consistent way to attach a user's username to a span.
func AttachUsernameToSpan(span *trace.Span, username string) {
	attachStringToSpan(span, keys.UsernameKey, username)
}

// AttachWebhookIDToSpan provides a consistent way to attach a webhook's ID to a span.
func AttachWebhookIDToSpan(span *trace.Span, webhookID uint64) {
	attachUint64ToSpan(span, keys.WebhookIDKey, webhookID)
}

// AttachRequestURIToSpan attaches a given URI to a span.
func AttachRequestURIToSpan(span *trace.Span, uri string) {
	attachStringToSpan(span, keys.RequestURIKey, uri)
}

// AttachSearchQueryToSpan attaches a given search query to a span.
func AttachSearchQueryToSpan(span *trace.Span, query string) {
	attachStringToSpan(span, keys.SearchQueryKey, query)
}
