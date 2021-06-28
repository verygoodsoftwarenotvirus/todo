package frontend

import (
	"context"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	serviceName string = "frontend_service"
)

type (
	// AuthService is a subset of the larger types.AuthService interface.
	AuthService interface {
		UserAttributionMiddleware(next http.Handler) http.Handler
		PermissionFilterMiddleware(permissions ...authorization.Permission) func(next http.Handler) http.Handler
		ServiceAdminMiddleware(next http.Handler) http.Handler

		AuthenticateUser(ctx context.Context, loginData *types.UserLoginInput) (*types.User, *http.Cookie, error)
		LogoutUser(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request, res http.ResponseWriter) error
	}

	// UsersService is a subset of the larger types.UsersService interface.
	UsersService interface {
		RegisterUser(ctx context.Context, registrationInput *types.UserRegistrationInput) (*types.UserCreationResponse, error)
		VerifyUserTwoFactorSecret(ctx context.Context, input *types.TOTPSecretVerificationInput) error
	}

	// Service serves HTML.
	Service interface {
		SetupRoutes(router routing.Router)
	}

	service struct {
		paymentManager            capitalism.PaymentManager
		usersService              UsersService
		logger                    logging.Logger
		tracer                    tracing.Tracer
		panicker                  panicking.Panicker
		authService               AuthService
		dataStore                 database.DataManager
		itemIDFetcher             func(*http.Request) uint64
		localizer                 *i18n.Localizer
		templateFuncMap           template.FuncMap
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		accountIDFetcher          func(*http.Request) uint64
		apiClientIDFetcher        func(*http.Request) uint64
		webhookIDFetcher          func(*http.Request) uint64
		useFakeData               bool
	}
)

// ProvideService builds a new Service.
func ProvideService(
	cfg *Config,
	logger logging.Logger,
	authService AuthService,
	usersService UsersService,
	dataStore database.DataManager,
	routeParamManager routing.RouteParamManager,
	paymentManager capitalism.PaymentManager,
) Service {
	svc := &service{
		useFakeData:               cfg.UseFakeData,
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		tracer:                    tracing.NewTracer(serviceName),
		panicker:                  panicking.NewProductionPanicker(),
		localizer:                 provideLocalizer(),
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		authService:               authService,
		usersService:              usersService,
		paymentManager:            paymentManager,
		dataStore:                 dataStore,
		apiClientIDFetcher:        routeParamManager.BuildRouteParamIDFetcher(logger, apiClientIDURLParamKey, "API client"),
		accountIDFetcher:          routeParamManager.BuildRouteParamIDFetcher(logger, accountIDURLParamKey, "account"),
		webhookIDFetcher:          routeParamManager.BuildRouteParamIDFetcher(logger, webhookIDURLParamKey, "webhook"),
		itemIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, itemIDURLParamKey, "item"),
		templateFuncMap: map[string]interface{}{
			"relativeTime":        relativeTime,
			"relativeTimeFromPtr": relativeTimeFromPtr,
		},
	}

	svc.templateFuncMap["translate"] = svc.getSimpleLocalizedString

	return svc
}
