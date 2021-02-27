package accounts

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
)

func buildTestService() *service {
	return &service{
		logger:                       logging.NewNonOperationalLogger(),
		accountCounter:               &mockmetrics.UnitCounter{},
		accountDataManager:           &mocktypes.AccountDataManager{},
		accountMembershipDataManager: &mocktypes.AccountUserMembershipDataManager{},
		accountIDFetcher:             func(req *http.Request) uint64 { return 0 },
		requestContextFetcher:        func(*http.Request) (*types.RequestContext, error) { return &types.RequestContext{}, nil },
		encoderDecoder:               mockencoding.NewMockEncoderDecoder(),
		tracer:                       tracing.NewTracer("test"),
	}
}

func TestProvideAccountsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, AccountIDURIParamKey, "account").Return(func(*http.Request) uint64 { return 0 })
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.AccountDataManager{},
			&mocktypes.AccountUserMembershipDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			rpm,
		)

		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
