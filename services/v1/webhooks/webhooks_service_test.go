package webhooks

import (
	"context"
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

func buildTestService() *Service {
	return &Service{
		logger:           noop.ProvideNoopLogger(),
		webhookCounter:   &mockmetrics.UnitCounter{},
		webhookDatabase:  &mockmodels.WebhookDataManager{},
		userIDFetcher:    func(req *http.Request) uint64 { return 0 },
		webhookIDFetcher: func(req *http.Request) uint64 { return 0 },
		encoderDecoder:   &mockencoding.EncoderDecoder{},
		eventManager:     newsman.NewNewsman(nil, nil),
	}
}

func TestProvideWebhooksService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectation := uint64(123)
		uc := &mockmetrics.UnitCounter{}
		uc.On("IncrementBy", expectation).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		dm := &mockmodels.WebhookDataManager{}
		dm.On("GetAllWebhooksCount", mock.Anything).Return(expectation, nil)

		actual, err := ProvideWebhooksService(
			context.Background(),
			noop.ProvideNoopLogger(),
			dm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			newsman.NewNewsman(nil, nil),
		)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		actual, err := ProvideWebhooksService(
			context.Background(),
			noop.ProvideNoopLogger(),
			&mockmodels.WebhookDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			newsman.NewNewsman(nil, nil),
		)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error setting count", func(t *testing.T) {
		expectation := uint64(123)
		uc := &mockmetrics.UnitCounter{}
		uc.On("IncrementBy", expectation).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		dm := &mockmodels.WebhookDataManager{}
		dm.On("GetAllWebhooksCount", mock.Anything).Return(expectation, errors.New("blah"))

		actual, err := ProvideWebhooksService(
			context.Background(),
			noop.ProvideNoopLogger(),
			dm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			newsman.NewNewsman(nil, nil),
		)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
