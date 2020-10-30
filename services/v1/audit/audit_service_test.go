package audit

import (
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService() *Service {
	return &Service{
		logger:                 noop.NewLogger(),
		auditLog:               &mockmodels.AuditLogEntryDataManager{},
		auditLogEntryIDFetcher: func(req *http.Request) uint64 { return 0 },
		sessionInfoFetcher:     func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
		encoderDecoder:         &mockencoding.EncoderDecoder{},
	}
}

func TestProvideAuditService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := ProvideAuditService(
			noop.NewLogger(),
			&mockmodels.AuditLogEntryDataManager{},
			func(req *http.Request) uint64 { return 0 },
			func(*http.Request) (*models.SessionInfo, error) { return &models.SessionInfo{}, nil },
			&mockencoding.EncoderDecoder{},
		)

		assert.NotNil(t, s)
	})
}
