package items

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService() *Service {
	return &Service{
		logger:             noop.NewLogger(),
		itemCounter:        &mockmetrics.UnitCounter{},
		itemDataManager:    &mockmodels.ItemDataManager{},
		itemIDFetcher:      func(req *http.Request) uint64 { return 0 },
		sessionInfoFetcher: func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
		encoderDecoder:     &mockencoding.EncoderDecoder{},
		search:             &mocksearch.IndexManager{},
	}
}

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		s, err := ProvideItemsService(
			noop.NewLogger(),
			&mockmodels.ItemDataManager{},
			&mockmodels.AuditLogDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			&mockencoding.EncoderDecoder{},
			ucp,
			&mocksearch.IndexManager{},
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})

	T.Run("with error providing unit counter", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		s, err := ProvideItemsService(
			noop.NewLogger(),
			&mockmodels.ItemDataManager{},
			&mockmodels.AuditLogDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			&mockencoding.EncoderDecoder{},
			ucp,
			&mocksearch.IndexManager{},
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
