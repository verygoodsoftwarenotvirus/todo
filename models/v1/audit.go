package models

import (
	"context"
	"net/http"
)

type (
	// AuditUpdateFieldDiff keeps track of what gets modified within audit reports
	AuditUpdateFieldDiff struct {
		FieldName string `json:"fieldName"`
		OldValue  string `json:"oldValue"`
		NewValue  string `json:"newValue"`
	}

	// AuditEvent represents an event we might want to log for audit purposes.
	AuditEvent struct {
		ID        uint64                 `json:"id"`
		EventType string                 `json:"name"`
		EventData []AuditUpdateFieldDiff `json:"changes"`
		CreatedOn uint64                 `json:"createdOn"`
	}

	// AuditEventList represents a list of items.
	AuditEventList struct {
		Pagination
		AuditEvents []AuditEvent `json:"items"`
	}

	// AuditEventCreationInput represents what a user could set as input for creating items.
	AuditEventCreationInput struct {
		Name          string `json:"name"`
		Details       string `json:"details"`
		BelongsToUser uint64 `json:"-"`
	}

	// AuditEventDataManager describes a structure capable of storing items permanently.
	AuditEventDataManager interface {
		GetAuditEvent(ctx context.Context, eventID uint64) (*AuditEvent, error)
		GetAllAuditEventsCount(ctx context.Context) (uint64, error)
		GetAllAuditEvents(ctx context.Context, resultChannel chan []AuditEvent) error
		GetAuditEvents(ctx context.Context, filter *QueryFilter) (*AuditEventList, error)
		CreateAuditEvent(ctx context.Context, input *AuditEventCreationInput) (*AuditEvent, error)
	}

	// AuditEventDataServer describes a structure capable of serving traffic related to items.
	AuditEventDataServer interface {
		CreationInputMiddleware(next http.Handler) http.Handler

		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
	}
)
