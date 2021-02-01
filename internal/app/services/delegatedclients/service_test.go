package delegatedclients

import (
	"errors"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	return &service{
		clientDataManager:      database.BuildMockDatabase(),
		logger:                 logging.NewNonOperationalLogger(),
		encoderDecoder:         &mockencoding.EncoderDecoder{},
		authenticator:          &mockauth.Authenticator{},
		urlClientIDExtractor:   func(req *http.Request) uint64 { return 0 },
		delegatedClientCounter: &mockmetrics.UnitCounter{},
		tracer:                 tracing.NewTracer(serviceName),
	}
}

func TestProvideDelegatedClientsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		mockDelegatedClientDataManager := &mocktypes.DelegatedClientDataManager{}

		s, err := ProvideDelegatedClientsService(
			logging.NewNonOperationalLogger(),
			mockDelegatedClientDataManager,
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockauth.Authenticator{},
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, nil
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, mockDelegatedClientDataManager)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		t.Parallel()
		mockDelegatedClientDataManager := &mocktypes.DelegatedClientDataManager{}

		s, err := ProvideDelegatedClientsService(
			logging.NewNonOperationalLogger(),
			mockDelegatedClientDataManager,
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockauth.Authenticator{},
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, errors.New("blah")
			},
		)

		assert.Error(t, err)
		assert.Nil(t, s)

		mock.AssertExpectationsForObjects(t, mockDelegatedClientDataManager)
	})
}
