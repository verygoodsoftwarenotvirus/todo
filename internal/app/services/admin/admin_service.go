package admin

import (
	"net/http"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "auth_service"
)

type (
	// service handles authentication service-wide.
	service struct {
		config             authservice.Config
		logger             logging.Logger
		authenticator      password.Authenticator
		userDB             types.AdminUserDataManager
		auditLog           types.AdminAuditManager
		encoderDecoder     encoding.EncoderDecoder
		sessionManager     *scs.SessionManager
		sessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)
		userIDFetcher      func(*http.Request) uint64
		tracer             tracing.Tracer
	}
)

// ProvideService builds a new AuthService.
func ProvideService(
	logger logging.Logger,
	cfg authservice.Config,
	authenticator password.Authenticator,
	userDataManager types.AdminUserDataManager,
	auditLog types.AdminAuditManager,
	sessionManager *scs.SessionManager,
	encoder encoding.EncoderDecoder,
) (types.AdminService, error) {
	svc := &service{
		logger:             logger.WithName(serviceName),
		encoderDecoder:     encoder,
		config:             cfg,
		userDB:             userDataManager,
		auditLog:           auditLog,
		authenticator:      authenticator,
		sessionManager:     sessionManager,
		sessionInfoFetcher: routeparams.SessionInfoFetcherFromRequestContext,
		userIDFetcher:      routeparams.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		tracer:             tracing.NewTracer(serviceName),
	}
	svc.sessionManager.Lifetime = cfg.CookieLifetime

	return svc, nil
}
