package models

import (
	"context"
	"net/http"
)

type (
	// AuditEvent represents an event we might want to log for audit purposes.
	AuditEvent struct {
		ID              uint64 `json:"id"`
		Type            string `json:"name"`
		CreatedOn       uint64 `json:"createdOn"`
		PerformedByUser uint64 `json:"performedByUser"`
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

	// AuditEventUpdateInput represents what a user could set as input for updating items.
	AuditEventUpdateInput struct {
		Name          string `json:"name"`
		Details       string `json:"details"`
		BelongsToUser uint64 `json:"-"`
	}

	// AuditUpdateFieldDiff
	AuditUpdateFieldDiff struct {
		Field    string
		OldValue string
		NewValue string
	}

	// AuditEventDataManager describes a structure capable of storing items permanently.
	AuditEventDataManager interface {
		GetAuditEvent(ctx context.Context, itemID, userID uint64) (*AuditEvent, error)
		GetAllAuditEventsCount(ctx context.Context) (uint64, error)
		GetAllAuditEvents(ctx context.Context, resultChannel chan []AuditEvent) error
		GetAuditEvents(ctx context.Context, filter *QueryFilter) (*AuditEventList, error)
		GetAuditEventsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]AuditEvent, error)
		CreateAuditEvent(ctx context.Context, input *AuditEventCreationInput) (*AuditEvent, error)
	}

	// AuditEventDataServer describes a structure capable of serving traffic related to items.
	AuditEventDataServer interface {
		CreationInputMiddleware(next http.Handler) http.Handler

		SearchHandler(res http.ResponseWriter, req *http.Request)
		ListHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
	}
)
