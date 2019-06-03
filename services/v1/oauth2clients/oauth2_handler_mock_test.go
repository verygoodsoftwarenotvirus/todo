package oauth2clients

import (
	"net/http"

	"github.com/stretchr/testify/mock"
	"gopkg.in/oauth2.v3"
	oauth2server "gopkg.in/oauth2.v3/server"
)

var _ oauth2Handler = (*mockOauth2Handler)(nil)

type mockOauth2Handler struct {
	mock.Mock
}

func (m *mockOauth2Handler) SetAllowGetAccessRequest(allowed bool) {
	m.Called(allowed)
}

func (m *mockOauth2Handler) SetClientAuthorizedHandler(handler oauth2server.ClientAuthorizedHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) SetClientScopeHandler(handler oauth2server.ClientScopeHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) SetClientInfoHandler(handler oauth2server.ClientInfoHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) SetUserAuthorizationHandler(handler oauth2server.UserAuthorizationHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) SetAuthorizeScopeHandler(handler oauth2server.AuthorizeScopeHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) SetResponseErrorHandler(handler oauth2server.ResponseErrorHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) SetInternalErrorHandler(handler oauth2server.InternalErrorHandler) {
	m.Called(handler)
}

func (m *mockOauth2Handler) ValidationBearerToken(req *http.Request) (oauth2.TokenInfo, error) {
	args := m.Called(req)
	return args.Get(0).(oauth2.TokenInfo), args.Error(1)
}

func (m *mockOauth2Handler) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	return m.Called(res, req).Error(0)
}

func (m *mockOauth2Handler) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	return m.Called(res, req).Error(0)
}
