package users

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	serviceName        = "users_service"
	topicName          = "users"
	counterDescription = "number of users managed by the users service"
	counterName        = metrics.CounterName(serviceName)
)

var (
	_ models.UserDataServer = (*Service)(nil)
)

type (
	// RequestValidator validates request.
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	secretGenerator interface {
		GenerateTwoFactorSecret() (string, error)
	}

	// UserIDFetcher fetches usernames from requests.
	UserIDFetcher func(*http.Request) uint64

	// Service handles our users.
	Service struct {
		cookieSecret        []byte
		userDataManager     models.UserDataManager
		authenticator       auth.Authenticator
		logger              logging.Logger
		encoderDecoder      encoding.EncoderDecoder
		userIDFetcher       UserIDFetcher
		userCounter         metrics.UnitCounter
		reporter            newsman.Reporter
		secretGenerator     secretGenerator
		userCreationEnabled bool
	}
)

// ProvideUsersService builds a new UsersService.
func ProvideUsersService(
	authSettings config.AuthSettings,
	logger logging.Logger,
	userDataManager models.UserDataManager,
	authenticator auth.Authenticator,
	userIDFetcher UserIDFetcher,
	encoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	reporter newsman.Reporter,
) (*Service, error) {
	if userIDFetcher == nil {
		return nil, errors.New("userIDFetcher must be provided")
	}

	counter, err := counterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		cookieSecret:        []byte(authSettings.CookieSecret),
		logger:              logger.WithName(serviceName),
		userDataManager:     userDataManager,
		authenticator:       authenticator,
		userIDFetcher:       userIDFetcher,
		encoderDecoder:      encoder,
		userCounter:         counter,
		reporter:            reporter,
		secretGenerator:     &standardSecretGenerator{},
		userCreationEnabled: authSettings.EnableUserSignup,
	}

	return svc, nil
}
