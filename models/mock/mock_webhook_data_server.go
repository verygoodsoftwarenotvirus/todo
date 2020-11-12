package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/stretchr/testify/mock"
)

var _ models.WebhookDataServer = (*WebhookDataServer)(nil)

// WebhookDataServer is a mocked models.WebhookDataServer for testing.
type WebhookDataServer struct {
	mock.Mock
}

// CreationInputMiddleware implements our interface requirements.
func (m *WebhookDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// UpdateInputMiddleware implements our interface requirements.
func (m *WebhookDataServer) UpdateInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ListHandler implements our interface requirements.
func (m *WebhookDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler implements our interface requirements.
func (m *WebhookDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler implements our interface requirements.
func (m *WebhookDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// UpdateHandler implements our interface requirements.
func (m *WebhookDataServer) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ArchiveHandler implements our interface requirements.
func (m *WebhookDataServer) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// AuditEntryHandler implements our interface requirements.
func (m *WebhookDataServer) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}
