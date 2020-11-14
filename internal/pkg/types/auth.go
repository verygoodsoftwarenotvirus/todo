package types

import (
	"bytes"
	"context"
	"encoding/gob"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

const (
	// SessionInfoKey is the non-string type we use for referencing SessionInfo structs.
	SessionInfoKey ContextKey = "session_info"

	// AdminAsUserKey is the non-string type we use for communicating whether a user is being impersonated.
	AdminAsUserKey ContextKey = "admin_as_user"
)

func init() {
	gob.Register(&SessionInfo{})
}

type (
	// SessionInfo represents what we encode in our authentication cookies.
	SessionInfo struct {
		UserID           uint64                        `json:"-"`
		UserIsAdmin      bool                          `json:"-"`
		AdminPermissions permissions.PermissionChecker `json:"-"`
	}

	// UserStatusResponse is what we encode when the frontend wants to check auth status.
	UserStatusResponse struct {
		Authenticated bool `json:"isAuthenticated"`
		IsAdmin       bool `json:"isAdmin"`
	}

	// AuthAuditManager describes a structure capable of .
	AuthAuditManager interface {
		LogCycleCookieSecretEvent(ctx context.Context, userID uint64)
		LogSuccessfulLoginEvent(ctx context.Context, userID uint64)
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
