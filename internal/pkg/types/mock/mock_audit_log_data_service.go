package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AuditLogEntryDataService = (*AuditLogDataService)(nil)

// AuditLogDataService is a mock types.AuditLogEntryDataService.
type AuditLogDataService struct {
	mock.Mock
}

// ListHandler implements our interface.
func (m *AuditLogDataService) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler implements our interface.
func (m *AuditLogDataService) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}
