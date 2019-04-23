package users

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const (

	// MiddlewareCtxKey is the context key we search for when interacting with user-related requests
	MiddlewareCtxKey models.ContextKey = "user_input"

	serviceName = "users_service"
)

type (
	// RequestValidator validates request
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	// Tracer is an arbitrary type alias we use for dependency injection
	Tracer opentracing.Tracer

	// Service handles our users
	Service struct {
		cookieSecret  []byte
		database      models.UserDataManager
		authenticator auth.Authenticator
		logger        logging.Logger
		tracer        Tracer
		encoder       encoding.EncoderDecoder
		userIDFetcher func(*http.Request) uint64
	}

	// UserIDFetcher fetches usernames from requests
	UserIDFetcher func(*http.Request) uint64
)

// ProvideUsersService builds a new UsersService
func ProvideUsersService(
	authSettings config.AuthSettings,
	logger logging.Logger,
	database database.Database,
	authenticator auth.Authenticator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
) *Service {
	if userIDFetcher == nil {
		panic("usernameFetcher must be provided")
	}

	us := &Service{
		cookieSecret:  []byte(authSettings.CookieSecret),
		logger:        logger.WithName(serviceName),
		database:      database,
		authenticator: authenticator,
		userIDFetcher: userIDFetcher,
		tracer:        tracing.ProvideTracer(serviceName),
		encoder:       encoder,
	}
	return us
}
