package mock

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.DelegatedClientDataService = (*DelegatedClientDataServer)(nil)

// DelegatedClientDataServer is a mocked types.DelegatedClientDataService for testing.
type DelegatedClientDataServer struct {
	mock.Mock
}

// ListHandler is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ArchiveHandler is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreationInputMiddleware is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// DelegatedClientInfoMiddleware is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) DelegatedClientInfoMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ExtractDelegatedClientFromRequest is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) ExtractDelegatedClientFromRequest(ctx context.Context, req *http.Request) (*types.DelegatedClient, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*types.DelegatedClient), args.Error(1)
}

// HandleAuthorizeRequest is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}

// HandleTokenRequest is the obligatory implementation for our interface.
func (m *DelegatedClientDataServer) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}

// AuditEntryHandler implements our interface requirements.
func (m *DelegatedClientDataServer) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// GetAuditLogEntriesForDelegatedClient is a mock function.
func (m *DelegatedClientDataServer) GetAuditLogEntriesForDelegatedClient(ctx context.Context, clientID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}
