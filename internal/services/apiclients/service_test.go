package apiclients

import (
	"net/http"
	"testing"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	return &service{
		apiClientDataManager:      database.BuildMockDatabase(),
		logger:                    logging.NewNonOperationalLogger(),
		encoderDecoder:            mockencoding.NewMockEncoderDecoder(),
		authenticator:             &authentication.MockAuthenticator{},
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		urlClientIDExtractor:      func(req *http.Request) uint64 { return 0 },
		apiClientCounter:          &mockmetrics.UnitCounter{},
		secretGenerator:           &random.MockGenerator{},
		tracer:                    tracing.NewTracer(serviceName),
		cfg:                       &config{},
	}
}

func TestProvideAPIClientsService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		mockAPIClientDataManager := &mocktypes.APIClientDataManager{}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.IsType(logging.NewNonOperationalLogger()),
			APIClientIDURIParamKey,
			"api client",
		).Return(func(*http.Request) uint64 { return 0 })

		s := ProvideAPIClientsService(
			logging.NewNonOperationalLogger(),
			mockAPIClientDataManager,
			&mocktypes.UserDataManager{},
			&authentication.MockAuthenticator{},
			mockencoding.NewMockEncoderDecoder(),
			func(counterName, description string) metrics.UnitCounter {
				return nil
			},
			rpm,
			&config{},
		)
		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, mockAPIClientDataManager, rpm)
	})
}
