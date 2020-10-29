package audit

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService() *Service {
	return &Service{
		logger:                 noop.NewLogger(),
		itemCounter:            &mockmetrics.UnitCounter{},
		auditLog:               &mockmodels.AuditLogEntryDataManager{},
		auditLogEntryIDFetcher: func(req *http.Request) uint64 { return 0 },
		sessionInfoFetcher:     func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
		encoderDecoder:         &mockencoding.EncoderDecoder{},
	}
}

func TestProvideItemsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		s, err := ProvideAuditService(
			noop.NewLogger(),
			&mockmodels.AuditLogEntryDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			ucp,
			&mockencoding.EncoderDecoder{},
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})

	T.Run("with error providing unit counter", func(t *testing.T) {
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		s, err := ProvideAuditService(
			noop.NewLogger(),
			&mockmodels.AuditLogEntryDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			ucp,
			&mockencoding.EncoderDecoder{},
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
