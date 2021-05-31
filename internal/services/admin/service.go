package admin

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/alexedwards/scs/v2"
)

const (
	serviceName = "auth_service"
)

type (
	// service handles passwords service-wide.
	service struct {
		config                    *authservice.Config
		logger                    logging.Logger
		authenticator             authentication.Authenticator
		userDB                    types.AdminUserDataManager
		auditLog                  types.AdminAuditManager
		encoderDecoder            encoding.ServerEncoderDecoder
		sessionManager            *scs.SessionManager
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		userIDFetcher             func(*http.Request) uint64
		tracer                    tracing.Tracer
	}
)

// ProvideService builds a new AuthService.
func ProvideService(
	logger logging.Logger,
	cfg *authservice.Config,
	authenticator authentication.Authenticator,
	userDataManager types.AdminUserDataManager,
	auditLog types.AdminAuditManager,
	sessionManager *scs.SessionManager,
	encoder encoding.ServerEncoderDecoder,
	routeParamManager routing.RouteParamManager,
) types.AdminService {
	svc := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		encoderDecoder:            encoder,
		config:                    cfg,
		userDB:                    userDataManager,
		auditLog:                  auditLog,
		authenticator:             authenticator,
		sessionManager:            sessionManager,
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		userIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		tracer:                    tracing.NewTracer(serviceName),
	}
	svc.sessionManager.Lifetime = cfg.Cookies.Lifetime

	return svc
}
