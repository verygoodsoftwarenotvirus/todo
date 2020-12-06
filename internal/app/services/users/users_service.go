package users

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName        = "users_service"
	counterDescription = "number of users managed by the users service"
	counterName        = metrics.CounterName(serviceName)
)

var _ types.UserDataService = (*Service)(nil)

var errNoUserIDFetcherProvided = errors.New("userIDFetcher must be provided")

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
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// Service handles our users.
	Service struct {
		userDataManager    types.UserDataManager
		auditLog           types.UserAuditManager
		authSettings       config.AuthSettings
		authenticator      password.Authenticator
		logger             logging.Logger
		encoderDecoder     encoding.EncoderDecoder
		userIDFetcher      UserIDFetcher
		sessionInfoFetcher SessionInfoFetcher
		userCounter        metrics.UnitCounter
		secretGenerator    secretGenerator
	}
)

// ProvideUsersService builds a new UsersService.
func ProvideUsersService(
	authSettings config.AuthSettings,
	logger logging.Logger,
	userDataManager types.UserDataManager,
	auditLog types.UserAuditManager,
	authenticator password.Authenticator,
	userIDFetcher UserIDFetcher,
	sessionInfoFetcher SessionInfoFetcher,
	encoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
) (*Service, error) {
	if userIDFetcher == nil {
		return nil, errNoUserIDFetcherProvided
	}

	counter, err := counterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		logger:             logger.WithName(serviceName),
		userDataManager:    userDataManager,
		auditLog:           auditLog,
		authenticator:      authenticator,
		userIDFetcher:      userIDFetcher,
		sessionInfoFetcher: sessionInfoFetcher,
		encoderDecoder:     encoder,
		authSettings:       authSettings,
		userCounter:        counter,
		secretGenerator:    &standardSecretGenerator{},
	}

	return svc, nil
}
