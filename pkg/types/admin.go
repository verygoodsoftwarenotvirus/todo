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

	// UserReputationUpdateInput represents what an admin User could provide as input for changing statuses.
	UserReputationUpdateInput struct {
		_ struct{}

		NewReputation accountStatus `json:"newReputation"`
		Reason        string        `json:"reason"`
		TargetUserID  string        `json:"targetUserID"`
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
		validation.Field(&i.TargetUserID, validation.Required),
	)
}
