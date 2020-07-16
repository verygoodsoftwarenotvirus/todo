package items

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search/mock"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func buildTestService() *Service {
	return &Service{
		logger:          noop.ProvideNoopLogger(),
		itemCounter:     &mockmetrics.UnitCounter{},
		itemDataManager: &mockmodels.ItemDataManager{},
		itemIDFetcher:   func(req *http.Request) uint64 { return 0 },
		userIDFetcher:   func(req *http.Request) uint64 { return 0 },
		encoderDecoder:  &mockencoding.EncoderDecoder{},
		search:          &mocksearch.MockIndexManager{},
		reporter:        nil,
	}
}

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		s, err := ProvideItemsService(
			noop.ProvideNoopLogger(),
			&mockmodels.ItemDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			nil,
			&mocksearch.MockIndexManager{},
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})

	T.Run("with error providing unit counter", func(t *testing.T) {
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		s, err := ProvideItemsService(
			noop.ProvideNoopLogger(),
			&mockmodels.ItemDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			nil,
			&mocksearch.MockIndexManager{},
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
