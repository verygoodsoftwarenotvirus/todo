package items

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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
)

func buildTestService() *service {
	return &service{
		logger:                logging.NewNonOperationalLogger(),
		itemCounter:           &mockmetrics.UnitCounter{},
		itemDataManager:       &mocktypes.ItemDataManager{},
		itemIDFetcher:         func(req *http.Request) uint64 { return 0 },
		requestContextFetcher: func(*http.Request) (*types.RequestContext, error) { return &types.RequestContext{}, nil },
		encoderDecoder:        mockencoding.NewMockEncoderDecoder(),
		search:                &mocksearch.IndexManager{},
		tracer:                tracing.NewTracer("test"),
	}
}

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var ucp metrics.UnitCounterProvider = func(counterName, description string) metrics.UnitCounter {
			return &mockmetrics.UnitCounter{}
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, ItemIDURIParamKey, "item").Return(func(*http.Request) uint64 { return 0 })

		s, err := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			search.Config{ItemsIndexPath: "example/path"},
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return &mocksearch.IndexManager{}, nil
			},
			rpm,
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
