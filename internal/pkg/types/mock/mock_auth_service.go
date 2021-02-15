package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AuthService = (*AuthService)(nil)

// AuthService is a mock types.AuthService.
type AuthService struct {
	mock.Mock
}

// StatusHandler implements our AuthService interface.
func (m *AuthService) StatusHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// LoginHandler implements our AuthService interface.
func (m *AuthService) LoginHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// LogoutHandler implements our AuthService interface.
func (m *AuthService) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CycleCookieSecretHandler implements our AuthService interface.
func (m *AuthService) CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// PASETOHandler implements our AuthService interface.
func (m *AuthService) PASETOHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CookieAuthenticationMiddleware implements our AuthService interface.
func (m *AuthService) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// UserAttributionMiddleware implements our AuthService interface.
func (m *AuthService) UserAttributionMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// AuthorizationMiddleware implements our AuthService interface.
func (m *AuthService) AuthorizationMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// AdminMiddleware implements our AuthService interface.
func (m *AuthService) AdminMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// UserLoginInputMiddleware implements our AuthService interface.
func (m *AuthService) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}
