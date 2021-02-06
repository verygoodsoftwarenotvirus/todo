package tracing

import (
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"

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

// AttachFilterToSpan provides a consistent way to attach a filter's info to a span.
func AttachFilterToSpan(span trace.Span, page uint64, limit uint8) {
	span.SetAttributes(
		label.String(keys.FilterPageKey, strconv.FormatUint(page, 10)),
		label.String(keys.FilterLimitKey, strconv.FormatUint(uint64(limit), 10)),
	)
}

// AttachAuditLogEntryIDToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryIDToSpan(span trace.Span, entryID uint64) {
	attachUint64ToSpan(span, keys.AuditLogEntryIDKey, entryID)
}

// AttachAuditLogEntryEventTypeToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryEventTypeToSpan(span trace.Span, eventType string) {
	attachStringToSpan(span, keys.AuditLogEntryEventTypeKey, eventType)
}

// AttachItemIDToSpan attaches an item ID to a given span.
func AttachItemIDToSpan(span trace.Span, itemID uint64) {
	attachUint64ToSpan(span, keys.ItemIDKey, itemID)
}

// AttachAccountIDToSpan provides a consistent way to attach an account's ID to a span.
func AttachAccountIDToSpan(span trace.Span, accountID uint64) {
	attachUint64ToSpan(span, keys.AccountIDKey, accountID)
}

// AttachAccountUserMembershipIDToSpan provides a consistent way to attach an account's ID to a span.
func AttachAccountUserMembershipIDToSpan(span trace.Span, accountUserMembershipID uint64) {
	attachUint64ToSpan(span, keys.AccountIDKey, accountUserMembershipID)
}

// AttachUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachUserIDToSpan(span trace.Span, userID uint64) {
	attachUint64ToSpan(span, keys.UserIDKey, userID)
}

// AttachPlanIDToSpan provides a consistent way to attach a plan's ID to a span.
func AttachPlanIDToSpan(span trace.Span, planID uint64) {
	attachUint64ToSpan(span, keys.AccountSubscriptionPlanIDKey, planID)
}

// AttachSessionInfoToSpan provides a consistent way to attach a SessionInfo object to a span.
func AttachSessionInfoToSpan(span trace.Span, userID uint64, userIsSiteAdmin bool) {
	attachUint64ToSpan(span, keys.UserIDKey, userID)
	attachBooleanToSpan(span, keys.UserIsAdminKey, userIsSiteAdmin)
}

// AttachDelegatedClientIDToSpan is a consistent way to attach an oauth2 client's ID to a span.
func AttachDelegatedClientIDToSpan(span trace.Span, oauth2ClientID uint64) {
	attachUint64ToSpan(span, keys.OAuth2ClientDatabaseIDKey, oauth2ClientID)
}

// AttachOAuth2ClientDatabaseIDToSpan is a consistent way to attach an oauth2 client's ID to a span.
func AttachOAuth2ClientDatabaseIDToSpan(span trace.Span, oauth2ClientID uint64) {
	attachUint64ToSpan(span, keys.OAuth2ClientDatabaseIDKey, oauth2ClientID)
}

// AttachOAuth2ClientIDToSpan is a consistent way to attach an oauth2 client's Client ID to a span.
func AttachOAuth2ClientIDToSpan(span trace.Span, clientID string) {
	attachStringToSpan(span, keys.OAuth2ClientIDKey, clientID)
}

// AttachUsernameToSpan provides a consistent way to attach a user's username to a span.
func AttachUsernameToSpan(span trace.Span, username string) {
	attachStringToSpan(span, keys.UsernameKey, username)
}

// AttachWebhookIDToSpan provides a consistent way to attach a webhook's ID to a span.
func AttachWebhookIDToSpan(span trace.Span, webhookID uint64) {
	attachUint64ToSpan(span, keys.WebhookIDKey, webhookID)
}

// AttachRequestURIToSpan attaches a given URI to a span.
func AttachRequestURIToSpan(span trace.Span, uri string) {
	attachStringToSpan(span, keys.RequestURIKey, uri)
}

// AttachSearchQueryToSpan attaches a given search query to a span.
func AttachSearchQueryToSpan(span trace.Span, query string) {
	attachStringToSpan(span, keys.SearchQueryKey, query)
}
