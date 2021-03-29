package accountsubscriptionplans

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
		logger:                             logging.NewNonOperationalLogger(),
		planCounter:                        &mockmetrics.UnitCounter{},
		accountSubscriptionPlanDataManager: &mocktypes.AccountSubscriptionPlanDataManager{},
		accountSubscriptionPlanIDFetcher:   func(req *http.Request) uint64 { return 0 },
		requestContextFetcher:              func(*http.Request) (*types.RequestContext, error) { return &types.RequestContext{}, nil },
		encoderDecoder:                     mockencoding.NewMockEncoderDecoder(),
		tracer:                             tracing.NewTracer("test"),
	}
}

func TestProvidePlansService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var ucp metrics.UnitCounterProvider = func(counterName, description string) metrics.UnitCounter {
			return &mockmetrics.UnitCounter{}
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.IsType(logging.NewNonOperationalLogger()), AccountSubscriptionPlanIDURIParamKey, "account subscription plan").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.AccountSubscriptionPlanDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			rpm,
		)

		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
