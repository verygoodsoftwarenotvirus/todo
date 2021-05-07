package frontend

import (
	"context"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	serviceName string = "frontends_service"
)

type (
	// AuthService is a subset of the larger types.AuthService interface.
	AuthService interface {
		UserLoginInputMiddleware(next http.Handler) http.Handler
		UserAttributionMiddleware(next http.Handler) http.Handler

		AuthenticateUser(ctx context.Context, loginData *types.UserLoginInput) (*types.User, *http.Cookie, error)
	}

	// UsersService is a subset of the larger types.UsersService interface.
	UsersService interface {
		UserRegistrationInputMiddleware(next http.Handler) http.Handler

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
) (*Service, error) {
	localizer, err := provideLocalizer()
	if err != nil {
		return nil, observability.PrepareError(err, logger, nil, "initializing localizer for frontend")
	}

	svc := &Service{
		useFakeData:               cfg.UseFakeData,
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		tracer:                    tracing.NewTracer(serviceName),
		panicker:                  panicking.NewProductionPanicker(),
		localizer:                 localizer,
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

	return svc, nil
}
