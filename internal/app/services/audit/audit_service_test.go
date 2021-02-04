package audit

import (
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
)

func buildTestService() *service {
	return &service{
		logger:                 logging.NewNonOperationalLogger(),
		auditLog:               &mocktypes.AuditLogEntryDataManager{},
		auditLogEntryIDFetcher: func(req *http.Request) uint64 { return 0 },
		sessionInfoFetcher:     func(*http.Request) (*types.SessionInfo, error) { return &types.SessionInfo{}, nil },
		encoderDecoder:         &mockencoding.EncoderDecoder{},
		tracer:                 tracing.NewTracer("test"),
	}
}

func TestProvideAuditService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := ProvideService(
			logging.NewNonOperationalLogger(),
			&mocktypes.AuditLogEntryDataManager{},
			&mockencoding.EncoderDecoder{},
			routeparams.NewRouteParamManager(),
		)

		assert.NotNil(t, s)
	})
}
