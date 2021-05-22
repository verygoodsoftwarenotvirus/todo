package users

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	random "gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
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

	// service handles our users.
	service struct {
		userDataManager           types.UserDataManager
		accountDataManager        types.AccountDataManager
		authSettings              *authentication.Config
		authenticator             passwords.Authenticator
		logger                    logging.Logger
		encoderDecoder            encoding.ServerEncoderDecoder
		userIDFetcher             func(*http.Request) uint64
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		userCounter               metrics.UnitCounter
		secretGenerator           random.Generator
		imageUploadProcessor      images.ImageUploadProcessor
		uploadManager             uploads.UploadManager
		tracer                    tracing.Tracer
	}
)

// ProvideUsersService builds a new UsersService.
func ProvideUsersService(
	authSettings *authentication.Config,
	logger logging.Logger,
	userDataManager types.UserDataManager,
	accountDataManager types.AccountDataManager,
	authenticator passwords.Authenticator,
	encoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	imageUploadProcessor images.ImageUploadProcessor,
	uploadManager uploads.UploadManager,
	routeParamManager routing.RouteParamManager,
) types.UserDataService {
	return &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		userDataManager:           userDataManager,
		accountDataManager:        accountDataManager,
		authenticator:             authenticator,
		userIDFetcher:             routeParamManager.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user"),
		sessionContextDataFetcher: authentication.FetchContextFromRequest,
		encoderDecoder:            encoder,
		authSettings:              authSettings,
		userCounter:               metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
		secretGenerator:           random.NewGenerator(logger),
		tracer:                    tracing.NewTracer(serviceName),
		imageUploadProcessor:      imageUploadProcessor,
		uploadManager:             uploadManager,
	}
}
