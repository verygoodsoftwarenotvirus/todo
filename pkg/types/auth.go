package types

import (
	"bytes"
	"context"
	"encoding/gob"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// SessionContextDataKey is the non-string type we use for referencing SessionContextData structs.
	SessionContextDataKey ContextKey = "session_context_data"
	// UserIDContextKey is the non-string type we use for referencing SessionContextData structs.
	UserIDContextKey ContextKey = "user_id"
	// AccountIDContextKey is the non-string type we use for referencing SessionContextData structs.
	AccountIDContextKey ContextKey = "account_id"
	// UserLoginInputContextKey is the non-string type we use for referencing SessionContextData structs.
	UserLoginInputContextKey ContextKey = "user_login_input"
	// UserRegistrationInputContextKey is the non-string type we use for referencing SessionContextData structs.
	UserRegistrationInputContextKey ContextKey = "user_registration_input"
)

func init() {
	gob.Register(&SessionContextData{})
}

type (
	// UserAccountMembershipInfo represents key information about an account membership.
	UserAccountMembershipInfo struct {
		AccountName  string   `json:"name"`
		AccountRoles []string `json:"-"`
		AccountID    uint64   `json:"accountID"`
	}

	// SessionContextData represents what we encode in our passwords cookies.
	SessionContextData struct {
		AccountPermissions map[uint64]authorization.AccountRolePermissionsChecker `json:"-"`
		Requester          RequesterInfo                                          `json:"-"`
		ActiveAccountID    uint64                                                 `json:"-"`
	}

	// RequesterInfo contains data relevant to the user making a request.
	RequesterInfo struct {
		ServicePermissions    authorization.ServiceRolePermissionChecker `json:"-"`
		Reputation            accountStatus                              `json:"-"`
		ReputationExplanation string                                     `json:"-"`
		UserID                uint64                                     `json:"-"`
	}

	// UserStatusResponse is what we encode when the frontend wants to check auth status.
	UserStatusResponse struct {
		UserReputation            accountStatus `json:"accountStatus,omitempty"`
		UserReputationExplanation string        `json:"reputationExplanation"`
		ActiveAccount             uint64        `json:"activeAccount,omitempty"`
		UserIsAuthenticated       bool          `json:"isAuthenticated"`
	}

	// ChangeActiveAccountInput represents what a User could set as input for switching accounts.
	ChangeActiveAccountInput struct {
		AccountID uint64 `json:"accountID"`
	}

	// PASETOCreationInput is used to create a PASETO.
	PASETOCreationInput struct {
		ClientID          string `json:"clientID"`
		AccountID         uint64 `json:"accountID"`
		RequestTime       int64  `json:"requestTime"`
		RequestedLifetime uint64 `json:"requestedLifetime,omitempty"`
	}

	// PASETOResponse is used to respond to a PASETO request.
	PASETOResponse struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expiresAt"`
	}

	// AuthService describes a structure capable of handling passwords and authorization requests.
	AuthService interface {
		StatusHandler(res http.ResponseWriter, req *http.Request)
		BeginSessionHandler(res http.ResponseWriter, req *http.Request)
		EndSessionHandler(res http.ResponseWriter, req *http.Request)
		CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request)
		PASETOHandler(res http.ResponseWriter, req *http.Request)
		ChangeActiveAccountHandler(res http.ResponseWriter, req *http.Request)

		PermissionFilterMiddleware(permissions ...authorization.Permission) func(next http.Handler) http.Handler
		CookieRequirementMiddleware(next http.Handler) http.Handler
		UserAttributionMiddleware(next http.Handler) http.Handler
		AuthorizationMiddleware(next http.Handler) http.Handler
		ServiceAdminMiddleware(next http.Handler) http.Handler

		AuthenticateUser(ctx context.Context, loginData *UserLoginInput) (*User, *http.Cookie, error)
		LogoutUser(ctx context.Context, sessionCtxData *SessionContextData, req *http.Request, res http.ResponseWriter) error
	}

	// AuthAuditManager describes a structure capable of auditing auth events.
	AuthAuditManager interface {
		LogCycleCookieSecretEvent(ctx context.Context, userID uint64)
		LogSuccessfulLoginEvent(ctx context.Context, userID uint64)
		LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64)
		LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64)
		LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64)
		LogLogoutEvent(ctx context.Context, userID uint64)
	}
)

var _ validation.ValidatableWithContext = (*ChangeActiveAccountInput)(nil)

// ValidateWithContext validates a ChangeActiveAccountInput.
func (x *ChangeActiveAccountInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.AccountID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*PASETOCreationInput)(nil)

// ValidateWithContext ensures our  provided UserLoginInput meets expectations.
func (i *PASETOCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.ClientID, validation.Required),
		validation.Field(&i.RequestTime, validation.Required),
	)
}

// AccountRolePermissionsChecker returns the relevant AccountRolePermissionsChecker.
func (x *SessionContextData) AccountRolePermissionsChecker() authorization.AccountRolePermissionsChecker {
	return x.AccountPermissions[x.ActiveAccountID]
}

// ServiceRolePermissionChecker returns the relevant ServiceRolePermissionChecker.
func (x *SessionContextData) ServiceRolePermissionChecker() authorization.ServiceRolePermissionChecker {
	return x.Requester.ServicePermissions
}

// ToBytes returns the gob encoded session info.
func (x *SessionContextData) ToBytes() []byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(x); err != nil {
		panic(err)
	}

	return b.Bytes()
}

// AttachToLogger provides a consistent way to attach a SessionContextData object to a logger.
func (x *SessionContextData) AttachToLogger(logger logging.Logger) logging.Logger {
	if x != nil {
		logger = logger.WithValue(keys.RequesterIDKey, x.Requester.UserID).
			WithValue(keys.ActiveAccountIDKey, x.ActiveAccountID)

		if x.Requester.ServicePermissions != nil {
			logger = logger.WithValue(keys.ServiceRoleKey, x.Requester.ServicePermissions.IsServiceAdmin())
		}
	}

	return logger
}
