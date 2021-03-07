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
		Context    AuditLogContext `json:"context"`
		ExternalID string          `json:"externalID"`
		EventType  string          `json:"eventType"`
		ID         uint64          `json:"id"`
		CreatedOn  uint64          `json:"createdOn"`
	}

	// AuditLogEntryList represents a list of items.
	AuditLogEntryList struct {
		Entries []*AuditLogEntry `json:"entries"`
		Pagination
	}

	// AuditLogEntryCreationInput represents what a User could set as input for creating items.
	AuditLogEntryCreationInput struct {
		Context   AuditLogContext `json:"context"`
		EventType string          `json:"eventType"`
	}

	// AuditLogEntrySQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AuditLogEntrySQLQueryBuilder interface {
		BuildGetAuditLogEntryQuery(entryID uint64) (query string, args []interface{})
		BuildGetAllAuditLogEntriesCountQuery() string
		BuildGetBatchOfAuditLogEntriesQuery(beginID, endID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesQuery(filter *QueryFilter) (query string, args []interface{})
		BuildCreateAuditLogEntryQuery(input *AuditLogEntryCreationInput) (query string, args []interface{})
	}

	// AuditLogEntryDataManager describes a structure capable of managing audit log entries.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, eventID uint64) (*AuditLogEntry, error)
		GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error)
		GetAllAuditLogEntries(ctx context.Context, resultChannel chan []*AuditLogEntry, bucketSize uint16) error
		GetAuditLogEntries(ctx context.Context, filter *QueryFilter) (*AuditLogEntryList, error)
	}

	// AuditLogEntryDataService describes a structure capable of serving traffic related to audit log entries.
	AuditLogEntryDataService interface {
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
