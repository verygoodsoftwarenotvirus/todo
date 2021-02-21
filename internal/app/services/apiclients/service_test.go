package apiclients

import (
	"errors"
	"net/http"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	return &service{
		apiClientDataManager: database.BuildMockDatabase(),
		logger:               logging.NewNonOperationalLogger(),
		encoderDecoder:       mockencoding.NewMockEncoderDecoder(),
		authenticator:        &mockauth.Authenticator{},
		urlClientIDExtractor: func(req *http.Request) uint64 { return 0 },
		apiClientCounter:     &mockmetrics.UnitCounter{},
		secretGenerator:      &mockSecretGenerator{},
		tracer:               tracing.NewTracer(serviceName),
	}
}

func TestProvideAPIClientsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		mockAPIClientDataManager := &mocktypes.APIClientDataManager{}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, APIClientIDURIParamKey, "api client").Return(func(*http.Request) uint64 { return 0 })

		s, err := ProvideAPIClientsService(
			logging.NewNonOperationalLogger(),
			mockAPIClientDataManager,
			&mocktypes.UserDataManager{},
			&mockauth.Authenticator{},
			mockencoding.NewMockEncoderDecoder(),
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, nil
			},
			rpm,
		)
		assert.NoError(t, err)
		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, mockAPIClientDataManager, rpm)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		t.Parallel()
		mockAPIClientDataManager := &mocktypes.APIClientDataManager{}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, APIClientIDURIParamKey, "api client").Return(func(*http.Request) uint64 { return 0 })

		s, err := ProvideAPIClientsService(
			logging.NewNonOperationalLogger(),
			mockAPIClientDataManager,
			&mocktypes.UserDataManager{},
			&mockauth.Authenticator{},
			mockencoding.NewMockEncoderDecoder(),
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, errors.New("blah")
			},
			rpm,
		)

		assert.Error(t, err)
		assert.Nil(t, s)

		mock.AssertExpectationsForObjects(t, mockAPIClientDataManager, rpm)
	})
}
