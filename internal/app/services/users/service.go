package users

import (
	"net/http"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	serviceName        = "users_service"
	counterDescription = "number of users managed by the users service"
	counterName        = metrics.CounterName("users")
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
		userDataManager       types.UserDataManager
		accountDataManager    types.AccountDataManager
		authSettings          *authservice.Config
		authenticator         authentication.Authenticator
		logger                logging.Logger
		encoderDecoder        encoding.ServerEncoderDecoder
		userIDFetcher         func(*http.Request) uint64
		requestContextFetcher func(*http.Request) (*types.RequestContext, error)
		userCounter           metrics.UnitCounter
		secretGenerator       secretGenerator
		imageUploadProcessor  images.ImageUploadProcessor
		uploadManager         uploads.UploadManager
		tracer                tracing.Tracer
	}
)

// ProvideUsersService builds a new UsersService.
func ProvideUsersService(
	authSettings *authservice.Config,
	logger logging.Logger,
	userDataManager types.UserDataManager,
	accountDataManager types.AccountDataManager,
	authenticator authentication.Authenticator,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	imageUploadProcessor images.ImageUploadProcessor,
	uploadManager uploads.UploadManager,
	routeParamManager routing.RouteParamManager,
) types.UserDataService {
	return &service{
		logger:                logging.EnsureLogger(logger).WithName(serviceName),
		userDataManager:       userDataManager,
		accountDataManager:    accountDataManager,
		authenticator:         authenticator,
		userIDFetcher:         routeParamManager.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		requestContextFetcher: routeParamManager.FetchContextFromRequest,
		encoderDecoder:        encoder,
		authSettings:          authSettings,
		userCounter:           metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		secretGenerator:       &standardSecretGenerator{},
		tracer:                tracing.NewTracer(serviceName),
		imageUploadProcessor:  imageUploadProcessor,
		uploadManager:         uploadManager,
	}
}
