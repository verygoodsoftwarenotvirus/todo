package items

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"
)

func buildTestService() *Service {
	return &Service{
		logger:         noop.ProvideNoopLogger(),
		itemCounter:    &mmetrics.UnitCounter{},
		itemDatabase:   &mmodels.ItemDataManager{},
		userIDFetcher:  func(req *http.Request) uint64 { return 0 },
		itemIDFetcher:  func(req *http.Request) uint64 { return 0 },
		encoderDecoder: &mencoding.EncoderDecoder{},
		reporter:       nil,
	}
}

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		uc := &mmetrics.UnitCounter{}
		expectation := uint64(123)

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		idm := &mmodels.ItemDataManager{}
		idm.On("GetAllItemsCount", mock.Anything).
			Return(expectation, nil)

		uc.On("IncrementBy", expectation).Return()

		s, err := ProvideItemsService(
			noop.ProvideNoopLogger(),
			idm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
			nil,
		)

		require.NotNil(t, s)
		require.NoError(t, err)
	})

	T.Run("with error providing unit counter", func(t *testing.T) {
		uc := &mmetrics.UnitCounter{}
		expectation := uint64(123)

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, errors.New("blah")
		}

		idm := &mmodels.ItemDataManager{}
		idm.On("GetAllItemsCount", mock.Anything).
			Return(expectation, nil)

		uc.On("IncrementBy", expectation).Return()

		s, err := ProvideItemsService(
			noop.ProvideNoopLogger(),
			idm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
			nil,
		)

		require.Nil(t, s)
		require.Error(t, err)
	})

	T.Run("with error fetching item count", func(t *testing.T) {
		uc := &mmetrics.UnitCounter{}
		expectation := uint64(123)

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		idm := &mmodels.ItemDataManager{}
		idm.On("GetAllItemsCount", mock.Anything).
			Return(expectation, errors.New("blah"))

		uc.On("IncrementBy", expectation).Return()

		s, err := ProvideItemsService(
			noop.ProvideNoopLogger(),
			idm,
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
			nil,
		)

		require.Nil(t, s)
		require.Error(t, err)
	})
}
