package webhooks

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
		logger:                logging.NewNonOperationalLogger(),
		webhookCounter:        &mockmetrics.UnitCounter{},
		webhookDataManager:    &mocktypes.WebhookDataManager{},
		requestContextFetcher: func(req *http.Request) (*types.RequestContext, error) { return &types.RequestContext{}, nil },
		webhookIDFetcher:      func(req *http.Request) uint64 { return 0 },
		encoderDecoder:        mockencoding.NewMockEncoderDecoder(),
		tracer:                tracing.NewTracer("test"),
	}
}

func TestProvideWebhooksService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, WebhookIDURIParamKey, "webhook").Return(func(*http.Request) uint64 { return 0 })

		actual := ProvideWebhooksService(
			logging.NewNonOperationalLogger(),
			&mocktypes.WebhookDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			rpm,
		)

		assert.NotNil(t, actual)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
