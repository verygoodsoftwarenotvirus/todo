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
	// UserRequestContext contains data relevant to the user making a request.
	UserRequestContext struct {
		Username                string                                        `json:"-"`
		ID                      uint64                                        `json:"-"`
		ActiveAccountID         uint64                                        `json:"-"`
		Status                  userReputation                                `json:"-"`
		AccountPermissionsMap   map[uint64]permissions.ServiceUserPermissions `json:"-"`
		ServiceAdminPermissions permissions.ServiceAdminPermissions           `json:"-"`
	}

	// RequestContext represents what we encode in our authentication cookies.
	RequestContext struct {
		User UserRequestContext `json:"-"`
	}

	// UserStatusResponse is what we encode when the frontend wants to check auth status.
	UserStatusResponse struct {
		UserIsAuthenticated      bool                                        `json:"isAuthenticated"`
		UserAccountStatus        userReputation                              `json:"accountStatus,omitempty"`
		AccountStatusExplanation string                                      `json:"statusExplanation,omitempty"`
		ServiceAdminPermissions  *permissions.ServiceAdminPermissionsSummary `json:"permissions,omitempty"`
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

		PermissionRestrictionMiddleware(p ...permissions.ServiceUserPermissions) func(next http.Handler) http.Handler
		CookieAuthenticationMiddleware(next http.Handler) http.Handler
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

// ToBytes returns the gob encoded session info.
func (x *RequestContext) ToBytes() []byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(x); err != nil {
		panic(err)
	}

	return b.Bytes()
}

// RequestContextFromUser produces a RequestContext object from a User's data.
func RequestContextFromUser(user *User, activeAccountID uint64, accountPermissionsMap map[uint64]permissions.ServiceUserPermissions) (*RequestContext, error) {
	if user == nil {
		return nil, errors.New("non-nil user required for request context")
	}

	if activeAccountID == 0 {
		return nil, errors.New("active account ID required for request context")
	}

	if accountPermissionsMap == nil {
		return nil, errors.New("non-nil permissions map required for request context")
	}

	reqCtx := &RequestContext{
		User: UserRequestContext{
			ID:                      user.ID,
			Username:                user.Username,
			Status:                  user.Reputation,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
			ActiveAccountID:         activeAccountID,
			AccountPermissionsMap:   accountPermissionsMap,
		},
	}

	return reqCtx, nil
}
