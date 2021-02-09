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
		ID            uint64        `json:"id"`
		ExternalID    string        `json:"externalID"`
		Name          string        `json:"name"`
		Description   string        `json:"description"`
		Price         uint32        `json:"price"`
		Period        time.Duration `json:"period"`
		CreatedOn     uint64        `json:"createdOn"`
		LastUpdatedOn *uint64       `json:"lastUpdatedOn"`
		ArchivedOn    *uint64       `json:"archivedOn"`
	}

	// AccountSubscriptionPlanList represents a list of account subscription plans.
	AccountSubscriptionPlanList struct {
		Pagination
		AccountSubscriptionPlans []*AccountSubscriptionPlan `json:"accountSubscriptionPlans"`
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

	// AccountSubscriptionPlanSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AccountSubscriptionPlanSQLQueryBuilder interface {
		BuildGetAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{})
		BuildGetAllAccountSubscriptionPlansCountQuery() string
		BuildGetAccountSubscriptionPlansQuery(filter *QueryFilter) (query string, args []interface{})
		BuildCreateAccountSubscriptionPlanQuery(input *AccountSubscriptionPlanCreationInput) (query string, args []interface{})
		BuildUpdateAccountSubscriptionPlanQuery(input *AccountSubscriptionPlan) (query string, args []interface{})
		BuildArchiveAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{})
	}

	// AccountSubscriptionPlanDataManager describes a structure capable of storing account subscription plans permanently.
	AccountSubscriptionPlanDataManager interface {
		GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*AccountSubscriptionPlan, error)
		GetAllAccountSubscriptionPlansCount(ctx context.Context) (uint64, error)
		GetAccountSubscriptionPlans(ctx context.Context, filter *QueryFilter) (*AccountSubscriptionPlanList, error)
		CreateAccountSubscriptionPlan(ctx context.Context, input *AccountSubscriptionPlanCreationInput) (*AccountSubscriptionPlan, error)
		UpdateAccountSubscriptionPlan(ctx context.Context, updated *AccountSubscriptionPlan, changedBy uint64, changes []FieldChangeSummary) error
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

// Update merges an PlanInput with an plan.
func (x *AccountSubscriptionPlan) Update(input *AccountSubscriptionPlanUpdateInput) []FieldChangeSummary {
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

// Validate validates a AccountSubscriptionPlanCreationInput.
func (x *AccountSubscriptionPlanCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}

// Validate validates a AccountSubscriptionPlanUpdateInput.
func (x *AccountSubscriptionPlanUpdateInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
