package users

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"github.com/opentracing/opentracing-go"
)

// MiddlewareCtxKey is the context key we search for when interacting with user-related requests
const MiddlewareCtxKey models.ContextKey = "user_input"

type (
	// RequestValidator validates request
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	// Tracer is an arbitrary type alias we use for dependency injection
	Tracer opentracing.Tracer

	// Service handles our users
	Service struct {
		cookieName      CookieName
		database        database.Database
		authenticator   auth.Enticator
		logger          logging.Logger
		tracer          Tracer
		encoder         encoding.ResponseEncoder
		usernameFetcher func(*http.Request) string
	}

	// UsernameFetcher fetches usernames from requests
	UsernameFetcher func(*http.Request) string

	// CookieName is an arbitrary type alias
	CookieName string
)

var (
	// Providers is what we provide for dependency injectors
	Providers = wire.NewSet(
		ProvideUserServiceTracer,
		ProvideUsersService,
	)
)

// ProvideUserServiceTracer wraps an opentracing Tracer
func ProvideUserServiceTracer() Tracer {
	return tracing.ProvideTracer("users-service")
}

// ProvideUsersService builds a new UsersService
func ProvideUsersService(
	cookieName CookieName,
	logger logging.Logger,
	database database.Database,
	authenticator auth.Enticator,
	usernameFetcher UsernameFetcher,
	tracer Tracer,
	encoder encoding.ResponseEncoder,
) *Service {
	if usernameFetcher == nil {
		panic("usernameFetcher must be provided")
	}
	us := &Service{
		logger:          logger.WithName("users_service"),
		cookieName:      cookieName,
		database:        database,
		authenticator:   authenticator,
		usernameFetcher: usernameFetcher,
		tracer:          tracer,
		encoder:         encoder,
	}
	return us
}
