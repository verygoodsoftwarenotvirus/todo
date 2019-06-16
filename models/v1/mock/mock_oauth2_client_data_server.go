package mock

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.OAuth2ClientDataServer = (*OAuth2ClientDataServer)(nil)

// OAuth2ClientDataServer describes a structure capable of serving traffic related to oauth2 clients
type OAuth2ClientDataServer struct {
	mock.Mock
}

// ListHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// DeleteHandler is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) DeleteHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
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
