package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// AdminService describes a structure capable of serving traffic related to users.
	AdminService interface {
		UserAccountStatusChangeHandler(res http.ResponseWriter, req *http.Request)

		AccountStatusUpdateInputMiddleware(next http.Handler) http.Handler
	}

	// AdminAuditManager describes a structure capable of managing audit entries for admin events.
	AdminAuditManager interface {
		LogUserBanEvent(ctx context.Context, banGiver, banReceiver uint64, reason string)
		LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string)
	}

	// AccountStatusUpdateInput represents what an admin User could provide as input for changing statuses.
	AccountStatusUpdateInput struct {
		TargetAccountID uint64            `json:"accountID"`
		NewStatus       userAccountStatus `json:"newStatus"`
		Reason          string            `json:"reason"`
	}

	// FrontendService serves static frontend files.
	FrontendService interface {
		StaticDir(staticFilesDirectory string) (http.HandlerFunc, error)
	}
)

// Validate ensures our struct is validatable.
func (i *AccountStatusUpdateInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.NewStatus, validation.Required),
		validation.Field(&i.Reason, validation.Required),
		validation.Field(&i.TargetAccountID, validation.Required, validation.Min(uint64(1))),
	)
}
