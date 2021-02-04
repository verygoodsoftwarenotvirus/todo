package accounts

import (
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
)

func buildTestService() *service {
	return &service{
		logger:             logging.NewNonOperationalLogger(),
		accountCounter:     &mockmetrics.UnitCounter{},
		accountDataManager: &mocktypes.AccountDataManager{},
		accountIDFetcher:   func(req *http.Request) uint64 { return 0 },
		sessionInfoFetcher: func(*http.Request) (*types.SessionInfo, error) { return &types.SessionInfo{}, nil },
		encoderDecoder:     &mockencoding.EncoderDecoder{},
		tracer:             tracing.NewTracer("test"),
	}
}

func TestProvideAccountsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		s, err := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.AccountDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockencoding.EncoderDecoder{},
			ucp,
			routeparams.NewRouteParamManager(),
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})

	T.Run("with error providing unit counter", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		s, err := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.AccountDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockencoding.EncoderDecoder{},
			ucp,
			routeparams.NewRouteParamManager(),
		)

		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
