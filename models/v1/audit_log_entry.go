package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
)

type (
	// AuditLogContext keeps track of what gets modified within audit reports
	AuditLogContext map[string]string

	// AuditLogEntry represents an event we might want to log for audit purposes.
	AuditLogEntry struct {
		ID        uint64          `json:"id"`
		EventType string          `json:"eventType"`
		Context   AuditLogContext `json:"context"`
		CreatedOn uint64          `json:"createdOn"`
	}

	// AuditLogEntryList represents a list of items.
	AuditLogEntryList struct {
		Pagination
		AuditLogEntries []AuditLogEntry `json:"auditLogEntries"`
	}

	// AuditLogEntryCreationInput represents what a user could set as input for creating items.
	AuditLogEntryCreationInput struct {
		EventType string          `json:"eventType"`
		Context   AuditLogContext `json:"context"`
	}

	// AuditLogEntryDataManager describes a structure capable of storing items permanently.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, eventID uint64) (*AuditLogEntry, error)
		GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error)
		GetAllAuditLogEntries(ctx context.Context, resultChannel chan []AuditLogEntry) error
		GetAuditLogEntries(ctx context.Context, filter *QueryFilter) (*AuditLogEntryList, error)
		CreateAuditLogEntry(ctx context.Context, input *AuditLogEntryCreationInput) (*AuditLogEntry, error)
	}

	// AuditLogEntryDataServer describes a structure capable of serving traffic related to items.
	AuditLogEntryDataServer interface {
		CreationInputMiddleware(next http.Handler) http.Handler

		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Make the Attrs struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (d AuditLogContext) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Make the Attrs struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (d *AuditLogContext) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}
