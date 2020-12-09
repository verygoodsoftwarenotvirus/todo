package oauth2clients

import (
	"net/http"

	oauth2 "github.com/go-oauth2/oauth2/v4"
	oauth2server "github.com/go-oauth2/oauth2/v4/server"
	"github.com/stretchr/testify/mock"
)

var _ oauth2Handler = (*mockOAuth2Handler)(nil)

type mockOAuth2Handler struct {
	mock.Mock
}

func (m *mockOAuth2Handler) SetAllowGetAccessRequest(allowed bool) {
	m.Called(allowed)
}

func (m *mockOAuth2Handler) SetClientAuthorizedHandler(handler oauth2server.ClientAuthorizedHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) SetClientScopeHandler(handler oauth2server.ClientScopeHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) SetClientInfoHandler(handler oauth2server.ClientInfoHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) SetUserAuthorizationHandler(handler oauth2server.UserAuthorizationHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) SetAuthorizeScopeHandler(handler oauth2server.AuthorizeScopeHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) SetResponseErrorHandler(handler oauth2server.ResponseErrorHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) SetInternalErrorHandler(handler oauth2server.InternalErrorHandler) {
	m.Called(handler)
}

func (m *mockOAuth2Handler) ValidationBearerToken(req *http.Request) (oauth2.TokenInfo, error) {
	args := m.Called(req)
	return args.Get(0).(oauth2.TokenInfo), args.Error(1)
}

func (m *mockOAuth2Handler) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	return m.Called(res, req).Error(0)
}

func (m *mockOAuth2Handler) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	return m.Called(res, req).Error(0)
}
