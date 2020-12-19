package users

import (
	"fmt"
	"net/http"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName        = "users_service"
	counterDescription = "number of users managed by the users service"
	counterName        = metrics.CounterName(serviceName)
)

var _ types.UserDataService = (*service)(nil)

type (
	// RequestValidator validates request.
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	secretGenerator interface {
		GenerateTwoFactorSecret() (string, error)
		GenerateSalt() ([]byte, error)
	}

	// service handles our users.
	service struct {
		userDataManager    types.UserDataManager
		auditLog           types.UserAuditManager
		authSettings       authservice.Config
		authenticator      password.Authenticator
		logger             logging.Logger
		encoderDecoder     encoding.EncoderDecoder
		userIDFetcher      func(*http.Request) uint64
		sessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)
		userCounter        metrics.UnitCounter
		secretGenerator    secretGenerator
	}
)

// ProvideUsersService builds a new UsersService.
func ProvideUsersService(
	authSettings authservice.Config,
	logger logging.Logger,
	userDataManager types.UserDataManager,
	auditLog types.UserAuditManager,
	authenticator password.Authenticator,
	encoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
) (types.UserDataService, error) {
	counter, err := counterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &service{
		logger:             logger.WithName(serviceName),
		userDataManager:    userDataManager,
		auditLog:           auditLog,
		authenticator:      authenticator,
		userIDFetcher:      routeparams.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		sessionInfoFetcher: routeparams.SessionInfoFetcherFromRequestContext,
		encoderDecoder:     encoder,
		authSettings:       authSettings,
		userCounter:        counter,
		secretGenerator:    &standardSecretGenerator{},
	}

	return svc, nil
}
