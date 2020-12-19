package webhooks

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService() *service {
	return &service{
		logger:             noop.NewLogger(),
		webhookCounter:     &mockmetrics.UnitCounter{},
		webhookDataManager: &mocktypes.WebhookDataManager{},
		sessionInfoFetcher: func(req *http.Request) (*types.SessionInfo, error) { return &types.SessionInfo{}, nil },
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
			&mocktypes.WebhookDataManager{},
			&mocktypes.AuditLogDataManager{},
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
			&mocktypes.WebhookDataManager{},
			&mocktypes.AuditLogDataManager{},
			&mockencoding.EncoderDecoder{},
			ucp,
		)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
