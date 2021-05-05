package types

import (
	"context"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// AccountSubscriptionPlan represents a plan.
	AccountSubscriptionPlan struct {
		LastUpdatedOn *uint64       `json:"lastUpdatedOn"`
		ArchivedOn    *uint64       `json:"archivedOn"`
		Description   string        `json:"description"`
		ExternalID    string        `json:"externalID"`
		Name          string        `json:"name"`
		ID            uint64        `json:"id"`
		Period        time.Duration `json:"period"`
		CreatedOn     uint64        `json:"createdOn"`
		Price         uint32        `json:"price"`
	}

	// AccountSubscriptionPlanList represents a list of account subscription plans.
	AccountSubscriptionPlanList struct {
		AccountSubscriptionPlans []*AccountSubscriptionPlan `json:"accountSubscriptionPlans"`
		Pagination
	}

	// AccountSubscriptionPlanCreationInput represents what a User could set as input for creating account subscription plans.
	AccountSubscriptionPlanCreationInput struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Price       uint32        `json:"price"`
		Period      time.Duration `json:"period"`
	}

	// AccountSubscriptionPlanUpdateInput represents what a User could set as input for updating account subscription plans.
	AccountSubscriptionPlanUpdateInput struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Price       uint32        `json:"price"`
		Period      time.Duration `json:"period"`
	}

	// AccountSubscriptionPlanDataManager describes a structure capable of storing account subscription plans permanently.
	AccountSubscriptionPlanDataManager interface {
		GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*AccountSubscriptionPlan, error)
		GetAllAccountSubscriptionPlansCount(ctx context.Context) (uint64, error)
		GetAccountSubscriptionPlans(ctx context.Context, filter *QueryFilter) (*AccountSubscriptionPlanList, error)
		CreateAccountSubscriptionPlan(ctx context.Context, input *AccountSubscriptionPlanCreationInput) (*AccountSubscriptionPlan, error)
		UpdateAccountSubscriptionPlan(ctx context.Context, updated *AccountSubscriptionPlan, changedBy uint64, changes []*FieldChangeSummary) error
		ArchiveAccountSubscriptionPlan(ctx context.Context, planID, archivedBy uint64) error
		GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]*AuditLogEntry, error)
	}

	// AccountSubscriptionPlanDataService describes a structure capable of serving traffic related to account subscription plans.
	AccountSubscriptionPlanDataService interface {
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

// Update merges an PlanInput with an account subscription plan.
func (x *AccountSubscriptionPlan) Update(input *AccountSubscriptionPlanUpdateInput) []*FieldChangeSummary {
	var out []*FieldChangeSummary

	if input.Name != "" && input.Name != x.Name {
		out = append(out, &FieldChangeSummary{
			FieldName: "Name",
			OldValue:  x.Name,
			NewValue:  input.Name,
		})

		x.Name = input.Name
	}

	return out
}

var _ validation.ValidatableWithContext = (*AccountSubscriptionPlanCreationInput)(nil)

// ValidateWithContext validates a AccountSubscriptionPlanCreationInput.
func (x *AccountSubscriptionPlanCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Description, validation.Required),
		validation.Field(&x.Price, validation.Required),
		validation.Field(&x.Period, validation.Required),
	)
}
