package users

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/pkg/errors"
)

const (
	// MiddlewareCtxKey is the context key we search for when interacting with user-related requests
	MiddlewareCtxKey models.ContextKey   = "user_input"
	counterName      metrics.CounterName = "users"
	topicName                            = "users"
	serviceName                          = "users_service"
)

type (
	// RequestValidator validates request
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	// Service handles our users
	Service struct {
		cookieSecret   []byte
		database       models.UserDataManager
		authenticator  auth.Authenticator
		logger         logging.Logger
		encoderDecoder encoding.EncoderDecoder
		userIDFetcher  UserIDFetcher
		userCounter    metrics.UnitCounter
		reporter       newsman.Reporter
	}

	// UserIDFetcher fetches usernames from requests
	UserIDFetcher func(*http.Request) uint64
)

// ProvideUsersService builds a new UsersService
func ProvideUsersService(
	ctx context.Context,
	authSettings config.AuthSettings,
	logger logging.Logger,
	database database.Database,
	authenticator auth.Authenticator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	newsman *newsman.Newsman,
) (*Service, error) {
	if userIDFetcher == nil {
		return nil, errors.New("userIDFetcher must be provided")
	}

	counter, err := counterProvider(counterName, "number of users managed by the users service")
	if err != nil {
		return nil, errors.Wrap(err, "error initializing counter")
	}

	userCount, err := database.GetUserCount(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fetching user count")
	}
	counter.IncrementBy(ctx, userCount)

	us := &Service{
		cookieSecret:   []byte(authSettings.CookieSecret),
		logger:         logger.WithName(serviceName),
		database:       database,
		authenticator:  authenticator,
		userIDFetcher:  userIDFetcher,
		encoderDecoder: encoder,
		userCounter:    counter,
		reporter:       newsman,
	}
	return us, nil
}
