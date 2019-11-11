package users

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// MiddlewareCtxKey is the context key we search for when interacting with user-related requests
	MiddlewareCtxKey models.ContextKey   = "user_input"
	counterName      metrics.CounterName = "users"
	topicName                            = "users"
	serviceName                          = "users_service"
)

var (
	_ models.UserDataServer = (*Service)(nil)
)

type (
	// RequestValidator validates request
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	// Service handles our users
	Service struct {
		cookieSecret        []byte
		database            database.Database
		authenticator       auth.Authenticator
		logger              logging.Logger
		encoderDecoder      encoding.EncoderDecoder
		userIDFetcher       UserIDFetcher
		userCounter         metrics.UnitCounter
		reporter            newsman.Reporter
		userCreationEnabled bool
	}

	// UserIDFetcher fetches usernames from requests
	UserIDFetcher func(*http.Request) uint64
)

// ProvideUsersService builds a new UsersService
func ProvideUsersService(
	ctx context.Context,
	authSettings config.AuthSettings,
	logger logging.Logger,
	db database.Database,
	authenticator auth.Authenticator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	reporter newsman.Reporter,
) (*Service, error) {
	if userIDFetcher == nil {
		return nil, errors.New("userIDFetcher must be provided")
	}

	counter, err := counterProvider(counterName, "number of users managed by the users service")
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	userCount, err := db.GetUserCount(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching user count: %w", err)
	}
	counter.IncrementBy(ctx, userCount)

	us := &Service{
		cookieSecret:        []byte(authSettings.CookieSecret),
		logger:              logger.WithName(serviceName),
		database:            db,
		authenticator:       authenticator,
		userIDFetcher:       userIDFetcher,
		encoderDecoder:      encoder,
		userCounter:         counter,
		reporter:            reporter,
		userCreationEnabled: authSettings.EnableUserSignup,
	}
	return us, nil
}
