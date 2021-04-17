package auth

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/securecookie"
)

const (
	serviceName         = "auth_service"
	userIDContextKey    = string(types.UserIDContextKey)
	accountIDContextKey = string(types.AccountIDContextKey)
	cookieErrorLogName  = "_COOKIE_CONSTRUCTION_ERROR_"
	cookieSecretSize    = 64
)

type (
	// cookieEncoderDecoder is a stand-in interface for gorilla/securecookie.
	cookieEncoderDecoder interface {
		Encode(name string, value interface{}) (string, error)
		Decode(name, value string, dst interface{}) error
	}

	// service handles authentication service-wide.
	service struct {
		config                    *Config
		logger                    logging.Logger
		authenticator             authentication.Authenticator
		userDataManager           types.UserDataManager
		auditLog                  types.AuthAuditManager
		apiClientManager          types.APIClientDataManager
		accountMembershipManager  types.AccountUserMembershipDataManager
		encoderDecoder            encoding.ServerEncoderDecoder
		cookieManager             cookieEncoderDecoder
		sessionManager            sessionManager
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		tracer                    tracing.Tracer
	}
)

// ProvideService builds a new AuthService.
func ProvideService(
	logger logging.Logger,
	cfg *Config,
	authenticator authentication.Authenticator,
	userDataManager types.UserDataManager,
	auditLog types.AuthAuditManager,
	apiClientsService types.APIClientDataManager,
	accountMembershipManager types.AccountUserMembershipDataManager,
	sessionManager *scs.SessionManager,
	encoder encoding.ServerEncoderDecoder,
	routeParamManager routing.RouteParamManager,
) (types.AuthService, error) {
	svc := &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		encoderDecoder:            encoder,
		config:                    cfg,
		userDataManager:           userDataManager,
		auditLog:                  auditLog,
		apiClientManager:          apiClientsService,
		accountMembershipManager:  accountMembershipManager,
		authenticator:             authenticator,
		sessionManager:            sessionManager,
		sessionContextDataFetcher: routeParamManager.FetchContextFromRequest,
		cookieManager: securecookie.New(
			securecookie.GenerateRandomKey(cookieSecretSize),
			[]byte(cfg.Cookies.SigningKey),
		),
		tracer: tracing.NewTracer(serviceName),
	}

	if _, err := svc.cookieManager.Encode(cfg.Cookies.Name, "blah"); err != nil {
		logger.WithValue("cookie_signing_key_length", len(cfg.Cookies.SigningKey)).Error(err, "building test cookie")
		return nil, fmt.Errorf("building test cookie: %w", err)
	}

	return svc, nil
}
