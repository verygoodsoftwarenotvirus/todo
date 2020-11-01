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

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName        = "users_service"
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
		GenerateSalt() ([]byte, error)
	}

	// UserIDFetcher fetches usernames from requests.
	UserIDFetcher func(*http.Request) uint64

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*models.SessionInfo, error)

	// Service handles our users.
	Service struct {
		cookieSecret        []byte
		userDataManager     models.UserDataManager
		auditLog            models.AuditLogDataManager
		authenticator       auth.Authenticator
		logger              logging.Logger
		encoderDecoder      encoding.EncoderDecoder
		userIDFetcher       UserIDFetcher
		sessionInfoFetcher  SessionInfoFetcher
		userCounter         metrics.UnitCounter
		secretGenerator     secretGenerator
		userCreationEnabled bool
	}
)

// ProvideUsersService builds a new UsersService.
func ProvideUsersService(
	authSettings config.AuthSettings,
	logger logging.Logger,
	userDataManager models.UserDataManager,
	auditLog models.AuditLogDataManager,
	authenticator auth.Authenticator,
	userIDFetcher UserIDFetcher,
	sessionInfoFetcher SessionInfoFetcher,
	encoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
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
		auditLog:            auditLog,
		authenticator:       authenticator,
		userIDFetcher:       userIDFetcher,
		sessionInfoFetcher:  sessionInfoFetcher,
		encoderDecoder:      encoder,
		userCounter:         counter,
		secretGenerator:     &standardSecretGenerator{},
		userCreationEnabled: authSettings.EnableUserSignup,
	}

	return svc, nil
}
