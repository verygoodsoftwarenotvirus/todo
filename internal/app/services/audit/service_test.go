package audit

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
)

func buildTestService() *service {
	return &service{
		logger:                 logging.NewNonOperationalLogger(),
		auditLog:               &mocktypes.AuditLogEntryDataManager{},
		auditLogEntryIDFetcher: func(req *http.Request) uint64 { return 0 },
		requestContextFetcher:  func(*http.Request) (*types.RequestContext, error) { return &types.RequestContext{}, nil },
		encoderDecoder:         mockencoding.NewMockEncoderDecoder(),
		tracer:                 tracing.NewTracer("test"),
	}
}

func TestProvideAuditService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, LogEntryURIParamKey, "audit log entry").Return(func(*http.Request) uint64 { return 0 })

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
