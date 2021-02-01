package delegatedclients

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

const (
	// creationMiddlewareCtxKey is a string alias for referring to Delegated client creation data.
	creationMiddlewareCtxKey types.ContextKey = "create_delegated_client"

	counterName        metrics.CounterName = "delegated_clients"
	counterDescription string              = "number of delegated clients managed by the delegated client service"
	serviceName        string              = "delegated_clients_service"
)

var _ types.DelegatedClientDataService = (*service)(nil)

type (
	// service manages our Delegated clients via HTTP.
	service struct {
		logger                 logging.Logger
		clientDataManager      types.DelegatedClientDataManager
		userDataManager        types.UserDataManager
		auditLog               types.DelegatedClientAuditManager
		authenticator          password.Authenticator
		encoderDecoder         encoding.EncoderDecoder
		urlClientIDExtractor   func(req *http.Request) uint64
		delegatedClientCounter metrics.UnitCounter
		initialized            bool
		tracer                 tracing.Tracer
	}
)

// ProvideDelegatedClientsService builds a new DelegatedClientsService.
func ProvideDelegatedClientsService(
	logger logging.Logger,
	clientDataManager types.DelegatedClientDataManager,
	userDataManager types.UserDataManager,
	auditLog types.DelegatedClientAuditManager,
	authenticator password.Authenticator,
	encoderDecoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
) (types.DelegatedClientDataService, error) {
	svc := &service{
		clientDataManager:    clientDataManager,
		auditLog:             auditLog,
		userDataManager:      userDataManager,
		logger:               logger.WithName(serviceName),
		encoderDecoder:       encoderDecoder,
		authenticator:        authenticator,
		urlClientIDExtractor: routeparams.BuildRouteParamIDFetcher(logger, DelegatedClientIDURIParamKey, "delegated client"),
		tracer:               tracing.NewTracer(serviceName),
	}

	var err error
	if svc.delegatedClientCounter, err = counterProvider(counterName, counterDescription); err != nil {
		return nil, fmt.Errorf("initializing counter: %w", err)
	}

	return svc, nil
}
