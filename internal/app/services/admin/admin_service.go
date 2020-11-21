package admin

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "auth_service"
)

type (
	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// UserIDFetcher is a function that fetches item IDs.
	UserIDFetcher func(*http.Request) uint64

	// Service handles authentication service-wide.
	Service struct {
		config             config.AuthSettings
		logger             logging.Logger
		authenticator      auth.Authenticator
		userDB             types.AdminUserDataManager
		auditLog           types.AdminAuditManager
		encoderDecoder     encoding.EncoderDecoder
		sessionManager     *scs.SessionManager
		sessionInfoFetcher SessionInfoFetcher
		userIDFetcher      UserIDFetcher
	}
)

// ProvideAdminService builds a new AuthService.
func ProvideAdminService(
	logger logging.Logger,
	cfg config.AuthSettings,
	authenticator auth.Authenticator,
	userDataManager types.AdminUserDataManager,
	auditLog types.AdminAuditManager,
	sessionManager *scs.SessionManager,
	encoder encoding.EncoderDecoder,
	sessionInfoFetcher SessionInfoFetcher,
	userIDFetcher UserIDFetcher,
) (*Service, error) {
	svc := &Service{
		logger:             logger.WithName(serviceName),
		encoderDecoder:     encoder,
		config:             cfg,
		userDB:             userDataManager,
		auditLog:           auditLog,
		authenticator:      authenticator,
		sessionManager:     sessionManager,
		sessionInfoFetcher: sessionInfoFetcher,
		userIDFetcher:      userIDFetcher,
	}
	svc.sessionManager.Lifetime = cfg.CookieLifetime

	return svc, nil
}
