package items

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService() *service {
	return &service{
		logger:          logging.NewNoopLogger(),
		itemCounter:     &mockmetrics.UnitCounter{},
		itemDataManager: &mocktypes.ItemDataManager{},
		itemIDFetcher:   func(req *http.Request) string { return "" },
		encoderDecoder:  mockencoding.NewMockEncoderDecoder(),
		search:          &mocksearch.IndexManager{},
		tracer:          tracing.NewTracer("test"),
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
		rpm.On(
			"BuildRouteParamStringIDFetcher",
			ItemIDURIParamKey,
		).Return(func(*http.Request) string { return "" })

		cfg := Config{SearchIndexPath: "example/path"}

		pp := &messagequeue.MockProducerProvider{}
		pp.On("ProviderProducer", cfg.PendingWritesTopicName).Return(&messagequeue.MockProducer{}, nil)

		s, err := ProvideService(
			logging.NewNoopLogger(),
			cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return &mocksearch.IndexManager{}, nil
			},
			rpm,
			pp,
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, rpm)
	})

	T.Run("with error providing index", func(t *testing.T) {
		t.Parallel()

		var ucp metrics.UnitCounterProvider = func(counterName, description string) metrics.UnitCounter {
			return &mockmetrics.UnitCounter{}
		}

		cfg := Config{SearchIndexPath: "example/path"}

		pp := &messagequeue.MockProducerProvider{}
		pp.On("ProviderProducer", cfg.PendingWritesTopicName).Return(&messagequeue.MockProducer{}, nil)

		s, err := ProvideService(
			logging.NewNoopLogger(),
			cfg,
			&mocktypes.ItemDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			func(path search.IndexPath, name search.IndexName, logger logging.Logger) (search.IndexManager, error) {
				return nil, errors.New("blah")
			},
			mockrouting.NewRouteParamManager(),
			pp,
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
