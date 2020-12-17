package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AuditLogDataService = (*AuditLogDataService)(nil)

// AuditLogDataService is a mock types.AuditLogDataService.
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
