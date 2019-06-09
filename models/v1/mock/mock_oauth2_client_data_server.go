package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.OAuth2ClientDataServer = (*OAuth2ClientDataServer)(nil)

// OAuth2ClientDataServer describes a structure capable of serving traffic related to oauth2 clients
type OAuth2ClientDataServer struct {
	mock.Mock
}

// List is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) List(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// Create is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) Create(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// Read is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) Read(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// Delete is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) Delete(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// CreationInputMiddleware is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called()
	return args.Get(0).(http.Handler)
}

// OAuth2ClientInfoMiddleware is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	args := m.Called()
	return args.Get(0).(http.Handler)
}

// RequestIsAuthenticated is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) RequestIsAuthenticated(req *http.Request) (*models.OAuth2Client, error) {
	args := m.Called()

	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// HandleAuthorizeRequest is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called()
	return args.Error(0)
}

// HandleTokenRequest is the obligatory implementation for our interface
func (m *OAuth2ClientDataServer) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	args := m.Called()
	return args.Error(0)
}
