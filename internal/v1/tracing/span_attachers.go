package tracing

import (
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const (
	auditLogEntryIDSpanAttachmentKey        = "audit_log_entry_id"
	auditLogEntryEventTypeSpanAttachmentKey = "event_type"
	itemIDSpanAttachmentKey                 = "item_id"
	userIDSpanAttachmentKey                 = "user_id"
	userIsAdminSpanAttachmentKey            = "is_admin"
	usernameSpanAttachmentKey               = "username"
	filterPageSpanAttachmentKey             = "filter_page"
	filterLimitSpanAttachmentKey            = "filter_limit"
	oauth2ClientDatabaseIDSpanAttachmentKey = "oauth2client_id"
	oauth2ClientIDSpanAttachmentKey         = "client_id"
	webhookIDSpanAttachmentKey              = "webhook_id"
	requestURISpanAttachmentKey             = "request_uri"
	searchQuerySpanAttachmentKey            = "search_query"
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
func AttachFilterToSpan(span *trace.Span, filter *models.QueryFilter) {
	if filter != nil && span != nil {
		span.AddAttributes(
			trace.StringAttribute(filterPageSpanAttachmentKey, strconv.FormatUint(filter.QueryPage(), 10)),
			trace.StringAttribute(filterLimitSpanAttachmentKey, strconv.FormatUint(uint64(filter.Limit), 10)),
		)
	}
}

// AttachAuditLogEntryIDToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryIDToSpan(span *trace.Span, entryID uint64) {
	attachUint64ToSpan(span, auditLogEntryIDSpanAttachmentKey, entryID)
}

// AttachAuditLogEntryEventTypeToSpan attaches an audit log entry ID to a given span.
func AttachAuditLogEntryEventTypeToSpan(span *trace.Span, eventType int) {
	attachIntToSpan(span, auditLogEntryEventTypeSpanAttachmentKey, eventType)
}

// AttachItemIDToSpan attaches an item ID to a given span.
func AttachItemIDToSpan(span *trace.Span, itemID uint64) {
	attachUint64ToSpan(span, itemIDSpanAttachmentKey, itemID)
}

// AttachUserIDToSpan provides a consistent way to attach a user's ID to a span.
func AttachUserIDToSpan(span *trace.Span, userID uint64) {
	attachUint64ToSpan(span, userIDSpanAttachmentKey, userID)
}

// AttachUserAdminStatusToSpan provides a consistent way to attach a user's admin status to a span.
func AttachUserAdminStatusToSpan(span *trace.Span, isAdmin bool) {
	attachBooleanToSpan(span, userIsAdminSpanAttachmentKey, isAdmin)
}

// AttachSessionInfoToSpan provides a consistent way to attach a SessionInfo object to a span.
func AttachSessionInfoToSpan(span *trace.Span, info models.SessionInfo) {
	attachUint64ToSpan(span, userIDSpanAttachmentKey, info.UserID)
	attachBooleanToSpan(span, userIsAdminSpanAttachmentKey, info.UserIsAdmin)
}

// AttachOAuth2ClientDatabaseIDToSpan is a consistent way to attach an oauth2 client's ID to a span.
func AttachOAuth2ClientDatabaseIDToSpan(span *trace.Span, oauth2ClientID uint64) {
	attachUint64ToSpan(span, oauth2ClientDatabaseIDSpanAttachmentKey, oauth2ClientID)
}

// AttachOAuth2ClientIDToSpan is a consistent way to attach an oauth2 client's Client ID to a span.
func AttachOAuth2ClientIDToSpan(span *trace.Span, clientID string) {
	attachStringToSpan(span, oauth2ClientIDSpanAttachmentKey, clientID)
}

// AttachUsernameToSpan provides a consistent way to attach a user's username to a span.
func AttachUsernameToSpan(span *trace.Span, username string) {
	attachStringToSpan(span, usernameSpanAttachmentKey, username)
}

// AttachWebhookIDToSpan provides a consistent way to attach a webhook's ID to a span.
func AttachWebhookIDToSpan(span *trace.Span, webhookID uint64) {
	attachUint64ToSpan(span, webhookIDSpanAttachmentKey, webhookID)
}

// AttachRequestURIToSpan attaches a given URI to a span.
func AttachRequestURIToSpan(span *trace.Span, uri string) {
	attachStringToSpan(span, requestURISpanAttachmentKey, uri)
}

// AttachSearchQueryToSpan attaches a given search query to a span.
func AttachSearchQueryToSpan(span *trace.Span, query string) {
	attachStringToSpan(span, searchQuerySpanAttachmentKey, query)
}
