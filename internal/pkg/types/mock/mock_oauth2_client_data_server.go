package mock

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.OAuth2ClientDataService = (*OAuth2ClientDataServer)(nil)

// OAuth2ClientDataServer is a mocked types.OAuth2ClientDataService for testing.
type OAuth2ClientDataServer struct {
	mock.Mock
}

// ListHandler is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ArchiveHandler is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreationInputMiddleware is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// OAuth2ClientInfoMiddleware is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ExtractOAuth2ClientFromRequest is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*types.OAuth2Client, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*types.OAuth2Client), args.Error(1)
}

// HandleAuthorizeRequest is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}

// HandleTokenRequest is the obligatory implementation for our interface.
func (m *OAuth2ClientDataServer) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}

// AuditEntryHandler implements our interface requirements.
func (m *OAuth2ClientDataServer) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// LogOAuth2ClientCreationEvent implements our interface.
func (m *AuditLogEntryDataManager) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	m.Called(ctx, client)
}

// LogOAuth2ClientArchiveEvent implements our interface.
func (m *AuditLogEntryDataManager) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	m.Called(ctx, userID, clientID)
}

// GetAuditLogEntriesForOAuth2Client is a mock function.
func (m *AuditLogEntryDataManager) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}
