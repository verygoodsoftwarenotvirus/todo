package frontend

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"net/http"
)

const (
	serviceName string = "frontends_service"
)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

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
	}

	// Service handles to-do list items.
	Service struct {
		logger       logging.Logger
		tracer       tracing.Tracer
		panicker     panicking.Panicker
		authService  AuthService
		usersService UsersService
		useFakes     bool
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(
	logger logging.Logger,
	authService AuthService,
	usersService UsersService,
) *Service {
	svc := &Service{
		useFakes:     true,
		logger:       logging.EnsureLogger(logger).WithName(serviceName),
		tracer:       tracing.NewTracer(serviceName),
		panicker:     panicking.NewProductionPanicker(),
		authService:  authService,
		usersService: usersService,
	}

	return svc
}
