package accounts

import (
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService() *service {
	return &service{
		logger:                       logging.NewNoopLogger(),
		accountCounter:               &mockmetrics.UnitCounter{},
		accountDataManager:           &mocktypes.AccountDataManager{},
		accountMembershipDataManager: &mocktypes.AccountUserMembershipDataManager{},
		accountIDFetcher:             func(req *http.Request) uint64 { return 0 },
		encoderDecoder:               mockencoding.NewMockEncoderDecoder(),
		tracer:                       tracing.NewTracer("test"),
	}
}

func TestProvideAccountsService(t *testing.T) {
	t.Parallel()

	var ucp metrics.UnitCounterProvider = func(counterName, description string) metrics.UnitCounter {
		return &mockmetrics.UnitCounter{}
	}

	l := logging.NewNoopLogger()

	rpm := mockrouting.NewRouteParamManager()
	rpm.On(
		"BuildRouteParamIDFetcher",
		mock.IsType(l), AccountIDURIParamKey, "account").Return(func(*http.Request) uint64 { return 0 })
	rpm.On(
		"BuildRouteParamIDFetcher",
		mock.IsType(l), UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

	s := ProvideService(
		logging.NewNoopLogger(),
		&mocktypes.AccountDataManager{},
		&mocktypes.AccountUserMembershipDataManager{},
		mockencoding.NewMockEncoderDecoder(),
		ucp,
		rpm,
	)

	assert.NotNil(t, s)

	mock.AssertExpectationsForObjects(t, rpm)
}
