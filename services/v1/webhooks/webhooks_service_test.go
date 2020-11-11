package webhooks

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
		logger:             noop.NewLogger(),
		webhookCounter:     &mockmetrics.UnitCounter{},
		webhookDataManager: &mockmodels.WebhookDataManager{},
		sessionInfoFetcher: func(req *http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
		webhookIDFetcher:   func(req *http.Request) uint64 { return 0 },
		encoderDecoder:     &mockencoding.EncoderDecoder{},
	}
}

func TestProvideWebhooksService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		actual, err := ProvideWebhooksService(
			noop.NewLogger(),
			&mockmodels.WebhookDataManager{},
			&mockmodels.AuditLogDataManager{},
			func(req *http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
		)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		actual, err := ProvideWebhooksService(
			noop.NewLogger(),
			&mockmodels.WebhookDataManager{},
			&mockmodels.AuditLogDataManager{},
			func(req *http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			ucp,
		)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
