package types

import (
	"context"
	"net/http"

	v "github.com/RussellLuo/validating/v2"
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
func (i *AccountStatusUpdateInput) Validate() error {
	err := v.Validate(v.Schema{
		v.F("newStatus", i.NewStatus): &userAccountStatusValidator{},
		v.F("reason", i.Reason):       &minimumStringLengthValidator{minLength: 1},
	})

	// for whatever reason, returning straight from v.Validate makes my tests fail /shrug
	if err != nil {
		return err
	}

	return nil
}
