package types

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// PlansSearchIndexName is the name of the index used to search through plans.
	PlansSearchIndexName search.IndexName = "plans"
)

type (
	// Plan represents a plan.
	Plan struct {
		ID            uint64        `json:"id"`
		Name          string        `json:"name"`
		Description   string        `json:"description"`
		Price         *uint32       `json:"price"`
		Period        time.Duration `json:"period"`
		CreatedOn     uint64        `json:"createdOn"`
		LastUpdatedOn *uint64       `json:"lastUpdatedOn"`
		ArchivedOn    *uint64       `json:"archivedOn"`
	}

	// PlanList represents a list of plans.
	PlanList struct {
		Pagination
		Plans []Plan `json:"plans"`
	}

	// PlanCreationInput represents what a User could set as input for creating plans.
	PlanCreationInput struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Price       *uint32       `json:"price"`
		Period      time.Duration `json:"period"`
	}

	// PlanUpdateInput represents what a User could set as input for updating plans.
	PlanUpdateInput struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Price       *uint32       `json:"price"`
		Period      time.Duration `json:"period"`
	}

	// PlanDataManager describes a structure capable of storing plans permanently.
	PlanDataManager interface {
		GetPlan(ctx context.Context, planID uint64) (*Plan, error)
		GetAllPlansCount(ctx context.Context) (uint64, error)
		GetPlans(ctx context.Context, filter *QueryFilter) (*PlanList, error)
		CreatePlan(ctx context.Context, input *PlanCreationInput) (*Plan, error)
		UpdatePlan(ctx context.Context, updated *Plan) error
		ArchivePlan(ctx context.Context, planID uint64) error
	}

	// PlanAuditManager describes a structure capable of .
	PlanAuditManager interface {
		GetAuditLogEntriesForPlan(ctx context.Context, planID uint64) ([]AuditLogEntry, error)
		LogPlanCreationEvent(ctx context.Context, plan *Plan)
		LogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []FieldChangeSummary)
		LogPlanArchiveEvent(ctx context.Context, userID, planID uint64)
	}

	// PlanDataService describes a structure capable of serving traffic related to plans.
	PlanDataService interface {
		CreationInputMiddleware(next http.Handler) http.Handler
		UpdateInputMiddleware(next http.Handler) http.Handler

		ListHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		UpdateHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
	}
)

// Update merges an PlanInput with an plan.
func (x *Plan) Update(input *PlanUpdateInput) []FieldChangeSummary {
	var out []FieldChangeSummary

	if input.Name != "" && input.Name != x.Name {
		out = append(out, FieldChangeSummary{
			FieldName: "Name",
			OldValue:  x.Name,
			NewValue:  input.Name,
		})

		x.Name = input.Name
	}

	return out
}

// Validate validates a PlanCreationInput.
func (x *PlanCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}

// Validate validates a PlanUpdateInput.
func (x *PlanUpdateInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
