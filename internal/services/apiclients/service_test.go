package apiclients

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	return &service{
		apiClientDataManager:      database.BuildMockDatabase(),
		logger:                    logging.NewNoopLogger(),
		encoderDecoder:            mockencoding.NewMockEncoderDecoder(),
		authenticator:             &authentication.MockAuthenticator{},
		sessionContextDataFetcher: authservice.FetchContextFromRequest,
		urlClientIDExtractor:      func(req *http.Request) string { return "" },
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
			"BuildRouteParamStringIDFetcher",
			APIClientIDURIParamKey,
		).Return(func(*http.Request) string { return "" })

		s := ProvideAPIClientsService(
			logging.NewNoopLogger(),
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
