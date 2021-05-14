package frontend

import (
	"context"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	panicking "gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
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
		AdminMiddleware(next http.Handler) http.Handler

		AuthenticateUser(ctx context.Context, loginData *types.UserLoginInput) (*types.User, *http.Cookie, error)
		LogoutUser(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request, res http.ResponseWriter) error
	}

	// UsersService is a subset of the larger types.UsersService interface.
	UsersService interface {
		RegisterUser(ctx context.Context, registrationInput *types.UserRegistrationInput) (*types.UserCreationResponse, error)
		VerifyUserTwoFactorSecret(ctx context.Context, input *types.TOTPSecretVerificationInput) error
	}

	// Service handles to-do list items.
	Service struct {
		logger                    logging.Logger
		tracer                    tracing.Tracer
		panicker                  panicking.Panicker
		routeParamManager         routing.RouteParamManager
		authService               AuthService
		usersService              UsersService
		dataStore                 database.DataManager
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		localizer                 *i18n.Localizer
		templateFuncMap           template.FuncMap
		useFakeData               bool
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(
	cfg *Config,
	logger logging.Logger,
	authService AuthService,
	usersService UsersService,
	dataStore database.DataManager,
	routeParamManager routing.RouteParamManager,
) *Service {
	svc := &Service{
		useFakeData:               cfg.UseFakeData,
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		tracer:                    tracing.NewTracer(serviceName),
		panicker:                  panicking.NewProductionPanicker(),
		localizer:                 provideLocalizer(),
		routeParamManager:         routeParamManager,
		sessionContextDataFetcher: routeParamManager.FetchContextFromRequest,
		authService:               authService,
		usersService:              usersService,
		dataStore:                 dataStore,
		templateFuncMap: map[string]interface{}{
			"relativeTime":        relativeTime,
			"relativeTimeFromPtr": relativeTimeFromPtr,
		},
	}

	svc.templateFuncMap["translate"] = svc.getSimpleLocalizedString

	return svc
}
