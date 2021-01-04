package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AccountDataService = (*AccountDataServer)(nil)

// AccountDataServer is a mocked types.AccountDataService for testing.
type AccountDataServer struct {
	mock.Mock
}

// CreationInputMiddleware implements our interface requirements.
func (m *AccountDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// UpdateInputMiddleware implements our interface requirements.
func (m *AccountDataServer) UpdateInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// SearchHandler implements our interface requirements.
func (m *AccountDataServer) SearchHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ListHandler implements our interface requirements.
func (m *AccountDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler implements our interface requirements.
func (m *AccountDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ExistenceHandler implements our interface requirements.
func (m *AccountDataServer) ExistenceHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler implements our interface requirements.
func (m *AccountDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// UpdateHandler implements our interface requirements.
func (m *AccountDataServer) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ArchiveHandler implements our interface requirements.
func (m *AccountDataServer) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// AuditEntryHandler implements our interface requirements.
func (m *AccountDataServer) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}
