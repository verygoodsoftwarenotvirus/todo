package admin

import (
	"net/http"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/v2"
)

const (
	serviceName = "auth_service"
)

type (
	// service handles authentication service-wide.
	service struct {
		config             *authservice.Config
		logger             logging.Logger
		authenticator      authentication.Authenticator
		userDB             types.AdminUserDataManager
		auditLog           types.AdminAuditManager
		encoderDecoder     encoding.HTTPResponseEncoder
		sessionManager     *scs.SessionManager
		sessionInfoFetcher func(*http.Request) (*types.RequestContext, error)
		userIDFetcher      func(*http.Request) uint64
		tracer             tracing.Tracer
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
	encoder encoding.HTTPResponseEncoder,
	routeParamManager routing.RouteParamManager,
) (types.AdminService, error) {
	svc := &service{
		logger:             logging.EnsureLogger(logger).WithName(serviceName),
		encoderDecoder:     encoder,
		config:             cfg,
		userDB:             userDataManager,
		auditLog:           auditLog,
		authenticator:      authenticator,
		sessionManager:     sessionManager,
		sessionInfoFetcher: routeParamManager.SessionInfoFetcherFromRequestContext,
		userIDFetcher:      routeParamManager.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		tracer:             tracing.NewTracer(serviceName),
	}
	svc.sessionManager.Lifetime = cfg.Cookies.Lifetime

	return svc, nil
}
