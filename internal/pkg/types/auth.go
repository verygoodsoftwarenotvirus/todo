package types

import (
	"bytes"
	"context"
	"encoding/gob"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

const (
	// SessionInfoKey is the non-string type we use for referencing SessionInfo structs.
	SessionInfoKey ContextKey = "session_info"
)

func init() {
	gob.Register(&SessionInfo{})
}

type (
	// SessionInfo represents what we encode in our authentication cookies.
	SessionInfo struct {
		Username          string
		UserID            uint64                                 `json:"-"`
		UserIsSiteAdmin   bool                                   `json:"-"`
		UserAccountStatus userReputation                         `json:"-"`
		AdminPermissions  permissions.SiteAdminPermissionChecker `json:"-"`
	}

	// UserStatusResponse is what we encode when the frontend wants to check auth status.
	UserStatusResponse struct {
		UserIsAuthenticated      bool                                     `json:"isAuthenticated"`
		UserIsAdmin              bool                                     `json:"isAdmin"`
		UserAccountStatus        userReputation                           `json:"accountStatus,omitempty"`
		AccountStatusExplanation string                                   `json:"statusExplanation,omitempty"`
		AdminPermissions         *permissions.SiteAdminPermissionsSummary `json:"permissions,omitempty"`
	}

	// AuthService describes a structure capable of handling authentication and authorization requests.
	AuthService interface {
		StatusHandler(res http.ResponseWriter, req *http.Request)
		LoginHandler(res http.ResponseWriter, req *http.Request)
		LogoutHandler(res http.ResponseWriter, req *http.Request)
		CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request)

		CookieAuthenticationMiddleware(next http.Handler) http.Handler
		UserAttributionMiddleware(next http.Handler) http.Handler
		AuthorizationMiddleware(next http.Handler) http.Handler
		AdminMiddleware(next http.Handler) http.Handler
		UserLoginInputMiddleware(next http.Handler) http.Handler
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

// ToBytes returns the gob encoded session info.
func (i *SessionInfo) ToBytes() []byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(i); err != nil {
		panic(err)
	}

	return b.Bytes()
}
