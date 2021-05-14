package audit

import (
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService() *service {
	return &service{
		logger:                 logging.NewNonOperationalLogger(),
		auditLog:               &mocktypes.AuditLogEntryDataManager{},
		auditLogEntryIDFetcher: func(req *http.Request) uint64 { return 0 },
		encoderDecoder:         mockencoding.NewMockEncoderDecoder(),
		tracer:                 tracing.NewTracer("test"),
	}
}

func TestProvideAuditService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.IsType(logging.NewNonOperationalLogger()), LogEntryURIParamKey, "audit log entry").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.AuditLogEntryDataManager{},
			mockencoding.NewMockEncoderDecoder(),
			rpm,
		)

		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
