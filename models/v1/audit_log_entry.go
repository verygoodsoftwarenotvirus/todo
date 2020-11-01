package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
)

type (
	// eventType is an enumertion-like string type
	eventType string

	// FieldChangeEvent represents a field that has changed in a given model's update
	FieldChangeEvent struct {
		FieldName string      `json:"fieldName"`
		OldValue  interface{} `json:"oldValue"`
		NewValue  interface{} `json:"newValue"`
	}

	// AuditLogContext keeps track of what gets modified within audit reports
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

	// AuditLogEntryDataManager describes a structure capable of storing items permanently.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, eventID uint64) (*AuditLogEntry, error)
		GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error)
		GetAllAuditLogEntries(ctx context.Context, resultChannel chan []AuditLogEntry) error
		GetAuditLogEntries(ctx context.Context, filter *QueryFilter) (*AuditLogEntryList, error)
		CreateAuditLogEntry(ctx context.Context, input *AuditLogEntryCreationInput)
	}

	// AuditLogEntryDataServer describes a structure capable of serving traffic related to items.
	AuditLogEntryDataServer interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Value implements the driver.Valuer interface.
func (d AuditLogContext) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner interface.
func (d *AuditLogContext) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

// Event Types

const (
	// CycleCookieSecretEventType events indicate an admin cycled the cookie secret.
	CycleCookieSecretEventType eventType = "cookie_secret_cycled"
	// SuccessfulLoginEventType events indicate a user successfully authenticated into the service via username + password + 2fa.
	SuccessfulLoginEventType eventType = "successful_login"
	// UnsuccessfulLoginBadPasswordEventType events indicate a user attempted to authenticate into the service, but failed because of an invalid password.
	UnsuccessfulLoginBadPasswordEventType eventType = "unsuccessful_login_bad_password"
	// UnsuccessfulLoginBad2FATokenEventType events indicate a user attempted to authenticate into the service, but failed because of a faulty two factor token.
	UnsuccessfulLoginBad2FATokenEventType eventType = "unsuccessful_login_bad_2fa_token"
	// LogoutEventType events indicate a user successfully logged out.
	LogoutEventType eventType = "user_logged_out"
	// ItemCreationEventType events indicate a user created an item.
	ItemCreationEventType eventType = "created_item"
	// ItemUpdateEventType events indicate a user updated an item.
	ItemUpdateEventType eventType = "updated_item"
	// ItemDeleteEventType events indicate a user deleted an item.
	ItemDeleteEventType eventType = "deleted_item"
)
