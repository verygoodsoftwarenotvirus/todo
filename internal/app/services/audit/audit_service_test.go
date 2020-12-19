package audit

import (
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService() *service {
	return &service{
		logger:                 noop.NewLogger(),
		auditLog:               &mocktypes.AuditLogDataManager{},
		auditLogEntryIDFetcher: func(req *http.Request) uint64 { return 0 },
		sessionInfoFetcher:     func(*http.Request) (*types.SessionInfo, error) { return &types.SessionInfo{}, nil },
		encoderDecoder:         &mockencoding.EncoderDecoder{},
	}
}

func TestProvideAuditService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := ProvideService(
			noop.NewLogger(),
			&mocktypes.AuditLogDataManager{},
			&mockencoding.EncoderDecoder{},
		)

		assert.NotNil(t, s)
	})
}
