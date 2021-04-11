package types

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

const (
	// RequestContextKey is the non-string type we use for referencing RequestContext structs.
	RequestContextKey ContextKey = "request_context"
	// UserIDContextKey is the non-string type we use for referencing RequestContext structs.
	UserIDContextKey ContextKey = "userID"
	// AccountIDContextKey is the non-string type we use for referencing RequestContext structs.
	AccountIDContextKey ContextKey = "accountID"
)

func init() {
	gob.Register(&RequestContext{})
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

	// RequestContext represents what we encode in our authentication cookies.
	RequestContext struct {
		AccountPermissionsMap AccountPermissionsMap `json:"-"`
		Requester             RequesterInfo         `json:"-"`
		ActiveAccountID       uint64                `json:"-"`
	}

	// RequesterInfo contains data relevant to the user making a request.
	RequesterInfo struct {
		Reputation             userReputation                     `json:"-"`
		ReputationExplanation  string                             `json:"-"`
		ID                     uint64                             `json:"-"`
		ServiceAdminPermission permissions.ServiceAdminPermission `json:"-"`
	}

	// UserStatusResponse is what we encode when the frontend wants to check auth status.
	UserStatusResponse struct {
		AccountPermissions             map[string]*UserAccountMembershipInfo       `json:"accountPermissions,omitempty"`
		ServiceAdminPermissionsSummary *permissions.ServiceAdminPermissionsSummary `json:"adminPermissions,omitempty"`
		UserReputation                 userReputation                              `json:"userReputation,omitempty"`
		UserReputationExplanation      string                                      `json:"reputationExplanation"`
		ActiveAccount                  uint64                                      `json:"activeAccount,omitempty"`
		UserIsAuthenticated            bool                                        `json:"isAuthenticated"`
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

	// AuthService describes a structure capable of handling authentication and authorization requests.
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
		UserLoginInputMiddleware(next http.Handler) http.Handler
		PASETOCreationInputMiddleware(next http.Handler) http.Handler
		ChangeActiveAccountInputMiddleware(next http.Handler) http.Handler
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

// Validate validates a ChangeActiveAccountInput.
func (x *ChangeActiveAccountInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.AccountID, validation.Required),
	)
}

// Validate ensures our  provided UserLoginInput meets expectations.
func (i *PASETOCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.ClientID, validation.Required),
		validation.Field(&i.RequestTime, validation.Required),
	)
}

// ToPermissionMapByAccountName returns an AccountPermissionsMap arranged by account name.
func (perms AccountPermissionsMap) ToPermissionMapByAccountName() map[string]*UserAccountMembershipInfo {
	out := map[string]*UserAccountMembershipInfo{}
	for _, v := range perms {
		out[v.AccountName] = v
	}

	return out
}

// ToBytes returns the gob encoded session info.
func (x *RequestContext) ToBytes() []byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(x); err != nil {
		panic(err)
	}

	return b.Bytes()
}

var (
	errNilUser          = errors.New("non-nil user required for request context")
	errZeroAccountID    = errors.New("active account ID required for request context")
	errNilPermissionMap = errors.New("non-nil permissions map required for request context")
)

// RequestContextFromUser produces a RequestContext object from a User's data.
func RequestContextFromUser(user *User, activeAccountID uint64, accountPermissionsMap AccountPermissionsMap) (*RequestContext, error) {
	if user == nil {
		return nil, errNilUser
	}

	if activeAccountID == 0 {
		return nil, errZeroAccountID
	}

	if accountPermissionsMap == nil {
		return nil, errNilPermissionMap
	}

	reqCtx := &RequestContext{
		Requester: RequesterInfo{
			ID:                     user.ID,
			Reputation:             user.Reputation,
			ReputationExplanation:  user.ReputationExplanation,
			ServiceAdminPermission: user.ServiceAdminPermission,
		},
		AccountPermissionsMap: accountPermissionsMap,
		ActiveAccountID:       activeAccountID,
	}

	return reqCtx, nil
}
