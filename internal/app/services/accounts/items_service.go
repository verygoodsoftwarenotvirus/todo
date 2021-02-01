package accounts

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	counterName        metrics.CounterName = "accounts"
	counterDescription string              = "the number of accounts managed by the accounts service"
	serviceName        string              = "accounts_service"
)

var _ types.AccountDataService = (*service)(nil)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// service handles to-do list accounts.
	service struct {
		logger             logging.Logger
		accountDataManager types.AccountDataManager
		auditLog           types.AccountAuditManager
		accountIDFetcher   func(*http.Request) uint64
		sessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)
		accountCounter     metrics.UnitCounter
		encoderDecoder     encoding.EncoderDecoder
		tracer             tracing.Tracer
	}
)

// ProvideService builds a new AccountsService.
func ProvideService(
	logger logging.Logger,
	accountDataManager types.AccountDataManager,
	auditLog types.AccountAuditManager,
	encoder encoding.EncoderDecoder,
	accountCounterProvider metrics.UnitCounterProvider,
) (types.AccountDataService, error) {
	accountCounter, err := accountCounterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("initializing counter: %w", err)
	}

	svc := &service{
		logger:             logger.WithName(serviceName),
		accountIDFetcher:   routeparams.BuildRouteParamIDFetcher(logger, AccountIDURIParamKey, "account"),
		sessionInfoFetcher: routeparams.SessionInfoFetcherFromRequestContext,
		accountDataManager: accountDataManager,
		auditLog:           auditLog,
		encoderDecoder:     encoder,
		accountCounter:     accountCounter,
		tracer:             tracing.NewTracer(serviceName),
	}

	return svc, nil
}
