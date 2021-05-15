package types

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

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

var (
	errNilUser          = errors.New("non-nil user required for session context data")
	errZeroAccountID    = errors.New("active account ID required for session context data")
	errNilPermissionMap = errors.New("non-nil permissions map required for session context data")
)

func init() {
	gob.Register(&SessionContextData{})
}

type (
	// UserAccountMembershipInfo represents key information about an account membership.
	UserAccountMembershipInfo struct {
		AccountName string                            `json:"name"`
		AccountID   uint64                            `json:"accountID"`
		Permissions permissions.ServiceUserPermission `json:"permissions"`
	}

	// AccountPermissionsMap maps accounts to membership info.
	AccountPermissionsMap map[uint64]*UserAccountMembershipInfo

	// SessionContextData represents what we encode in our passwords cookies.
	SessionContextData struct {
		AccountPermissionsMap AccountPermissionsMap `json:"-"`
		Requester             RequesterInfo         `json:"-"`
		ActiveAccountID       uint64                `json:"-"`
	}

	// FrontendUserAccountMembershipInfo represents key information about an account membership to the frontend.
	FrontendUserAccountMembershipInfo struct {
		UserAccountMembershipInfo
		Permissions permissions.ServiceUserPermissionsSummary `json:"permissions"`
	}

	// RequesterInfo contains data relevant to the user making a request.
	RequesterInfo struct {
		Reputation             userReputation                     `json:"-"`
		ReputationExplanation  string                             `json:"-"`
		ID                     uint64                             `json:"-"`
		ServiceAdminPermission permissions.ServiceAdminPermission `json:"-"`
		RequiresPasswordChange bool                               `json:"-"`
	}

	// UserStatusResponse is what we encode when the frontend wants to check auth status.
	UserStatusResponse struct {
		AccountPermissions             map[string]*FrontendUserAccountMembershipInfo `json:"accountPermissions,omitempty"`
		ServiceAdminPermissionsSummary *permissions.ServiceAdminPermissionsSummary   `json:"adminPermissions,omitempty"`
		UserReputation                 userReputation                                `json:"userReputation,omitempty"`
		UserReputationExplanation      string                                        `json:"reputationExplanation"`
		ActiveAccount                  uint64                                        `json:"activeAccount,omitempty"`
		UserIsAuthenticated            bool                                          `json:"isAuthenticated"`
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
		LoginHandler(res http.ResponseWriter, req *http.Request)
		LogoutHandler(res http.ResponseWriter, req *http.Request)
		CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request)
		PASETOHandler(res http.ResponseWriter, req *http.Request)
		ChangeActiveAccountHandler(res http.ResponseWriter, req *http.Request)

		PermissionRestrictionMiddleware(perms ...permissions.ServiceUserPermission) func(next http.Handler) http.Handler
		CookieRequirementMiddleware(next http.Handler) http.Handler
		UserAttributionMiddleware(next http.Handler) http.Handler
		AuthorizationMiddleware(next http.Handler) http.Handler
		AdminMiddleware(next http.Handler) http.Handler

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

// ToPermissionMapByAccountName returns an AccountPermissionsMap arranged by account name.
func (perms AccountPermissionsMap) ToPermissionMapByAccountName() map[string]*FrontendUserAccountMembershipInfo {
	out := map[string]*FrontendUserAccountMembershipInfo{}
	for _, v := range perms {
		out[v.AccountName] = &FrontendUserAccountMembershipInfo{
			UserAccountMembershipInfo: *v,
			Permissions:               v.Permissions.Summary(),
		}
	}

	return out
}

// ToBytes returns the gob encoded session info.
func (x *SessionContextData) ToBytes() []byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(x); err != nil {
		panic(err)
	}

	return b.Bytes()
}

// SessionContextDataFromUser produces a SessionContextData object from a User's data.
func SessionContextDataFromUser(user *User, activeAccountID uint64, accountPermissionsMap AccountPermissionsMap) (*SessionContextData, error) {
	if user == nil {
		return nil, errNilUser
	}

	if activeAccountID == 0 {
		return nil, errZeroAccountID
	}

	if accountPermissionsMap == nil {
		return nil, errNilPermissionMap
	}

	sessionCtxData := &SessionContextData{
		Requester: RequesterInfo{
			ID:                     user.ID,
			Reputation:             user.Reputation,
			ReputationExplanation:  user.ReputationExplanation,
			ServiceAdminPermission: user.ServiceAdminPermission,
			RequiresPasswordChange: user.RequiresPasswordChange,
		},
		AccountPermissionsMap: accountPermissionsMap,
		ActiveAccountID:       activeAccountID,
	}

	return sessionCtxData, nil
}
