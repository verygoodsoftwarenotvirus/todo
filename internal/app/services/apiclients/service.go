package apiclients

import (
	"crypto/rand"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

const (
	// creationMiddlewareCtxKey is a string alias for referring to API client creation data.
	creationMiddlewareCtxKey types.ContextKey = "create_api_client"

	counterName        metrics.CounterName = "api_clients"
	counterDescription string              = "number of API clients managed by the API client service"
	serviceName        string              = "api_clients_service"
)

var _ types.APIClientDataService = (*service)(nil)

type (
	// service manages our API clients via HTTP.
	service struct {
		logger                    logging.Logger
		apiClientDataManager      types.APIClientDataManager
		userDataManager           types.UserDataManager
		authenticator             authentication.Authenticator
		encoderDecoder            encoding.ServerEncoderDecoder
		urlClientIDExtractor      func(req *http.Request) uint64
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		apiClientCounter          metrics.UnitCounter
		secretGenerator           secretGenerator
		tracer                    tracing.Tracer
	}
)

// ProvideAPIClientsService builds a new APIClientsService.
func ProvideAPIClientsService(
	logger logging.Logger,
	clientDataManager types.APIClientDataManager,
	userDataManager types.UserDataManager,
	authenticator authentication.Authenticator,
	encoderDecoder encoding.ServerEncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) types.APIClientDataService {
	return &service{
		apiClientDataManager:      clientDataManager,
		userDataManager:           userDataManager,
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		encoderDecoder:            encoderDecoder,
		authenticator:             authenticator,
		secretGenerator:           &standardSecretGenerator{},
		sessionContextDataFetcher: routeParamManager.FetchContextFromRequest,
		urlClientIDExtractor:      routeParamManager.BuildRouteParamIDFetcher(logger, APIClientIDURIParamKey, "api client"),
		tracer:                    tracing.NewTracer(serviceName),
		apiClientCounter:          metrics.EnsureUnitCounter(counterProvider, logger, counterName, counterDescription),
	}
}
