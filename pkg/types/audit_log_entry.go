package types

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
)

type (
	// FieldChangeSummary represents a field that has changed in a given model's update.
	FieldChangeSummary struct {
		OldValue  interface{} `json:"oldValue"`
		NewValue  interface{} `json:"newValue"`
		FieldName string      `json:"fieldName"`
	}

	// AuditLogContext keeps track of what gets modified within audit reports.
	AuditLogContext map[string]interface{}

	// AuditLogEntry represents an event we might want to log for audit purposes.
	AuditLogEntry struct {
		Context   AuditLogContext `json:"context"`
		EventType string          `json:"eventType"`
		ID        string          `json:"id"`
		CreatedOn uint64          `json:"createdOn"`
	}

	// AuditLogEntryList represents a list of audit log entries.
	AuditLogEntryList struct {
		Entries []*AuditLogEntry `json:"entries"`
		Pagination
	}

	// AuditLogEntryCreationInput represents input for creating audit log entries.
	AuditLogEntryCreationInput struct {
		ID        string          `json:"-"`
		Context   AuditLogContext `json:"context"`
		EventType string          `json:"eventType"`
	}

	// AuditLogEntryDataManager describes a structure capable of managing audit log entries.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, eventID string) (*AuditLogEntry, error)
		GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error)
		GetAuditLogEntries(ctx context.Context, filter *QueryFilter) (*AuditLogEntryList, error)
	}

	// AuditLogEntryDataService describes a structure capable of serving traffic related to audit log entries.
	AuditLogEntryDataService interface {
		ListHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Value implements the driver.Valuer interface.
func (c AuditLogContext) Value() (driver.Value, error) {
	return json.Marshal(c)
}

var errByteAssertionFailed = errors.New("type assertion to []byte failed")

// Scan implements the sql.Scanner interface.
func (c *AuditLogContext) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errByteAssertionFailed
	}

	return json.Unmarshal(b, &c)
}
