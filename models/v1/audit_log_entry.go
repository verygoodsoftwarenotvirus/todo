package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
)

type (
	// eventType is an enum alias.
	eventType int

	// FieldChangeSummary represents a field that has changed in a given model's update.
	FieldChangeSummary struct {
		FieldName string      `json:"fieldName"`
		OldValue  interface{} `json:"oldValue"`
		NewValue  interface{} `json:"newValue"`
	}

	// AuditLogContext keeps track of what gets modified within audit reports.
	AuditLogContext map[string]interface{}

	// AuditLogEntry represents an event we might want to log for audit purposes.
	AuditLogEntry struct {
		ID        uint64          `json:"id"`
		EventType eventType       `json:"eventType"`
		Context   AuditLogContext `json:"context"`
		CreatedOn uint64          `json:"createdOn"`
	}

	// AuditLogEntryList represents a list of items.
	AuditLogEntryList struct {
		Pagination
		Entries []AuditLogEntry `json:"entries"`
	}

	// AuditLogEntryCreationInput represents what a user could set as input for creating items.
	AuditLogEntryCreationInput struct {
		EventType eventType       `json:"eventType"`
		Context   AuditLogContext `json:"context"`
	}

	// AuditLogDataManager describes a structure capable of storing items permanently.
	AuditLogDataManager interface {
		GetAuditLogEntry(ctx context.Context, eventID uint64) (*AuditLogEntry, error)
		GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error)
		GetAllAuditLogEntries(ctx context.Context, resultChannel chan []AuditLogEntry) error
		GetAuditLogEntries(ctx context.Context, filter *QueryFilter) (*AuditLogEntryList, error)

		LogCycleCookieSecretEvent(ctx context.Context, userID uint64)
		LogSuccessfulLoginEvent(ctx context.Context, userID uint64)
		LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64)
		LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64)
		LogLogoutEvent(ctx context.Context, userID uint64)
		LogItemCreationEvent(ctx context.Context, item *Item)
		LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []FieldChangeSummary)
		LogItemArchiveEvent(ctx context.Context, userID, itemID uint64)
		LogOAuth2ClientCreationEvent(ctx context.Context, client *OAuth2Client)
		LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64)
		LogWebhookCreationEvent(ctx context.Context, webhook *Webhook)
		LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []FieldChangeSummary)
		LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64)
		LogUserCreationEvent(ctx context.Context, user *User)
		LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64)
		LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64)
		LogUserUpdatePasswordEvent(ctx context.Context, userID uint64)
		LogUserArchiveEvent(ctx context.Context, userID uint64)
	}

	// AuditLogDataServer describes a structure capable of serving traffic related to items.
	AuditLogDataServer interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Value implements the driver.Valuer interface.
func (d AuditLogContext) Value() (driver.Value, error) {
	return json.Marshal(d)
}

var errByteAssertionFailed = errors.New("type assertion to []byte failed")

// Scan implements the sql.Scanner interface.
func (d *AuditLogContext) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errByteAssertionFailed
	}

	return json.Unmarshal(b, &d)
}

// Event Types

const (
	// CycleCookieSecretEvent events indicate an admin cycled the cookie secret.
	CycleCookieSecretEvent = iota
	// SuccessfulLoginEvent events indicate a user successfully authenticated into the service via username + password + 2fa.
	SuccessfulLoginEvent
	// UnsuccessfulLoginBadPasswordEvent events indicate a user attempted to authenticate into the service, but failed because of an invalid password.
	UnsuccessfulLoginBadPasswordEvent
	// UnsuccessfulLoginBad2FATokenEvent events indicate a user attempted to authenticate into the service, but failed because of a faulty two factor token.
	UnsuccessfulLoginBad2FATokenEvent
	// LogoutEvent events indicate a user successfully logged out.
	LogoutEvent
	// ItemCreationEvent events indicate a user created an item.
	ItemCreationEvent
	// ItemUpdateEvent events indicate a user updated an item.
	ItemUpdateEvent
	// ItemArchiveEvent events indicate a user deleted an item.
	ItemArchiveEvent
	// OAuth2ClientCreationEvent events indicate a user created an item.
	OAuth2ClientCreationEvent
	// OAuth2ClientArchiveEvent events indicate a user deleted an item.
	OAuth2ClientArchiveEvent
	// WebhookCreationEvent events indicate a user created an item.
	WebhookCreationEvent
	// WebhookUpdateEvent events indicate a user updated an item.
	WebhookUpdateEvent
	// WebhookArchiveEvent events indicate a user deleted an item.
	WebhookArchiveEvent
	// UserCreationEvent events indicate a user was created.
	UserCreationEvent
	// UserVerifyTwoFactorSecretEvent events indicate a user was created.
	UserVerifyTwoFactorSecretEvent
	// UserUpdateTwoFactorSecretEvent events indicate a user updated their two factor secret.
	UserUpdateTwoFactorSecretEvent
	// UserUpdatePasswordEvent events indicate a user updated their two factor secret.
	UserUpdatePasswordEvent
	// UserArchiveEvent events indicate a user was archived.
	UserArchiveEvent
)
