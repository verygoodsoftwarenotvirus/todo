package items

import (
	"context"
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func buildTestService() *Service {
	return &Service{
		logger:         noop.ProvideNoopLogger(),
		itemCounter:    &mockmetrics.UnitCounter{},
		itemDatabase:   &mockmodels.ItemDataManager{},
		userIDFetcher:  func(req *http.Request) uint64 { return 0 },
		itemIDFetcher:  func(req *http.Request) uint64 { return 0 },
		encoderDecoder: &mockencoding.EncoderDecoder{},
		reporter:       nil,
	}
}

func TestProvideItemsService(T *testing.T) {
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

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetAllItemsCount", mock.Anything).Return(expectation, nil)

		s, err := ProvideItemsService(
			context.Background(),
			noop.ProvideNoopLogger(),
			idm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			nil,
		)

		require.NotNil(t, s)
		require.NoError(t, err)
	})

	T.Run("with error providing unit counter", func(t *testing.T) {
		expectation := uint64(123)
		uc := &mockmetrics.UnitCounter{}
		uc.On("IncrementBy", expectation).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, errors.New("blah")
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetAllItemsCount", mock.Anything).Return(expectation, nil)

		s, err := ProvideItemsService(
			context.Background(),
			noop.ProvideNoopLogger(),
			idm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			nil,
		)

		require.Nil(t, s)
		require.Error(t, err)
	})

	T.Run("with error fetching item count", func(t *testing.T) {
		expectation := uint64(123)
		uc := &mockmetrics.UnitCounter{}
		uc.On("IncrementBy", expectation).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetAllItemsCount", mock.Anything).Return(expectation, errors.New("blah"))

		s, err := ProvideItemsService(
			context.Background(),
			noop.ProvideNoopLogger(),
			idm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
			nil,
		)

		require.Nil(t, s)
		require.Error(t, err)
	})
}
