package webhooks

import (
	"github.com/pkg/errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	mockman "gitlab.com/verygoodsoftwarenotvirus/newsman/mock"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
)

func buildTestService() *Service {
	return &Service{
		logger:           noop.ProvideNoopLogger(),
		webhookCounter:   &mmetrics.UnitCounter{},
		webhookDatabase:  &mmodels.WebhookDataManager{},
		userIDFetcher:    func(req *http.Request) uint64 { return 0 },
		webhookIDFetcher: func(req *http.Request) uint64 { return 0 },
		encoderDecoder:   &mencoding.EncoderDecoder{},
		newsman:          nil,
	}
}

var _ eventManager = (*eventMan)(nil)

type eventMan struct {
	mock.Mock

	*mockman.Reporter
}

func (m *eventMan) TuneIn(l newsman.Listener) {
	m.Called(l)
}

func TestProvideWebhooksService(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expectation := uint64(123)
		uc := &mmetrics.UnitCounter{}
		uc.On("IncrementBy", expectation).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		dm := &mmodels.WebhookDataManager{}
		dm.On("GetAllWebhooksCount", mock.Anything).
			Return(expectation, nil)

		actual, err := ProvideWebhooksService(
			noop.ProvideNoopLogger(),
			dm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
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
			noop.ProvideNoopLogger(),
			&mmodels.WebhookDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
			newsman.NewNewsman(nil, nil),
		)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error setting count", func(t *testing.T) {
		expectation := uint64(123)
		uc := &mmetrics.UnitCounter{}
		uc.On("IncrementBy", expectation).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		dm := &mmodels.WebhookDataManager{}
		dm.On("GetAllWebhooksCount", mock.Anything).
			Return(expectation, errors.New("blah"))

		actual, err := ProvideWebhooksService(
			noop.ProvideNoopLogger(),
			dm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
			newsman.NewNewsman(nil, nil),
		)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

}
