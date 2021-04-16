package apiclients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/random"
	"net/http"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	return &service{
		apiClientDataManager:      database.BuildMockDatabase(),
		logger:                    logging.NewNonOperationalLogger(),
		encoderDecoder:            mockencoding.NewMockEncoderDecoder(),
		authenticator:             &mockauth.Authenticator{},
		sessionContextDataFetcher: chi.NewRouteParamManager().FetchContextFromRequest,
		urlClientIDExtractor:      func(req *http.Request) uint64 { return 0 },
		apiClientCounter:          &mockmetrics.UnitCounter{},
		secretGenerator:           &random.MockGenerator{},
		tracer:                    tracing.NewTracer(serviceName),
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
			mock.IsType(logging.NewNonOperationalLogger()), APIClientIDURIParamKey, "api client").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideAPIClientsService(
			logging.NewNonOperationalLogger(),
			mockAPIClientDataManager,
			&mocktypes.UserDataManager{},
			&mockauth.Authenticator{},
			mockencoding.NewMockEncoderDecoder(),
			func(counterName, description string) metrics.UnitCounter {
				return nil
			},
			rpm,
		)
		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, mockAPIClientDataManager, rpm)
	})
}
