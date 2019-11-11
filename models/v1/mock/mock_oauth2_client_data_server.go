package mock

import (
	"context"
	"net/http"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.OAuth2ClientDataServer = (*OAuth2ClientDataServer)(nil)

// OAuth2ClientDataServer is a mocked models.OAuth2ClientDataServer for testing
type OAuth2ClientDataServer struct {
	mock.Mock
}

// ListHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) ListHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// CreateHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) CreateHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// ReadHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) ReadHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// ArchiveHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) ArchiveHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// CreationInputMiddleware is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// OAuth2ClientInfoMiddleware is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ExtractOAuth2ClientFromRequest is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*models.OAuth2Client, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// HandleAuthorizeRequest is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}

// HandleTokenRequest is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}
