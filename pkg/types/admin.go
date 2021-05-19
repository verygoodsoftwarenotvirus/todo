package types

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// AdminService describes a structure capable of serving traffic related to users.
	AdminService interface {
		UserReputationChangeHandler(res http.ResponseWriter, req *http.Request)
	}

	// AdminAuditManager describes a structure capable of managing audit entries for admin events.
	AdminAuditManager interface {
		LogUserBanEvent(ctx context.Context, banGiver, banReceiver uint64, reason string)
		LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string)
	}

	// UserReputationUpdateInput represents what an admin User could provide as input for changing statuses.
	UserReputationUpdateInput struct {
		NewReputation accountStatus `json:"newReputation"`
		Reason        string        `json:"reason"`
		TargetUserID  uint64        `json:"targetUserID"`
	}

	// FrontendService serves static frontend files.
	FrontendService interface {
		StaticDir(ctx context.Context, staticFilesDirectory string) (http.HandlerFunc, error)
	}
)

var _ validation.ValidatableWithContext = (*UserReputationUpdateInput)(nil)

// ValidateWithContext ensures our struct is validatable.
func (i *UserReputationUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.NewReputation, validation.Required),
		validation.Field(&i.Reason, validation.Required),
		validation.Field(&i.TargetUserID, validation.Required, validation.Min(uint64(1))),
	)
}
