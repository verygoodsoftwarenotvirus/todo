package mock

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

// AuthService is a mock types.AuthService.
type AuthService struct {
	mock.Mock
}

// StatusHandler satisfies our interface contract.
func (m *AuthService) StatusHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(req, res)
}

// LoginHandler satisfies our interface contract.
func (m *AuthService) LoginHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(req, res)
}

// LogoutHandler satisfies our interface contract.
func (m *AuthService) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(req, res)
}

// CycleCookieSecretHandler satisfies our interface contract.
func (m *AuthService) CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(req, res)
}

// PASETOHandler satisfies our interface contract.
func (m *AuthService) PASETOHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(req, res)
}

// ChangeActiveAccountHandler satisfies our interface contract.
func (m *AuthService) ChangeActiveAccountHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(req, res)
}

// PermissionRestrictionMiddleware satisfies our interface contract.
func (m *AuthService) PermissionRestrictionMiddleware(perms ...permissions.ServiceUserPermission) func(next http.Handler) http.Handler {
	return m.Called(perms).Get(0).(func(next http.Handler) http.Handler)
}

// CookieRequirementMiddleware satisfies our interface contract.
func (m *AuthService) CookieRequirementMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// UserAttributionMiddleware satisfies our interface contract.
func (m *AuthService) UserAttributionMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// AuthorizationMiddleware satisfies our interface contract.
func (m *AuthService) AuthorizationMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// AdminMiddleware satisfies our interface contract.
func (m *AuthService) AdminMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// UserLoginInputMiddleware satisfies our interface contract.
func (m *AuthService) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// PASETOCreationInputMiddleware satisfies our interface contract.
func (m *AuthService) PASETOCreationInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// ChangeActiveAccountInputMiddleware satisfies our interface contract.
func (m *AuthService) ChangeActiveAccountInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// AuthenticateUser satisfies our interface contract.
func (m *AuthService) AuthenticateUser(ctx context.Context, loginData *types.UserLoginInput) (*types.User, *http.Cookie, error) {
	returnValues := m.Called(ctx, loginData)

	return returnValues.Get(0).(*types.User), returnValues.Get(1).(*http.Cookie), returnValues.Error(2)
}

// LogoutUser satisfies our interface contract.
func (m *AuthService) LogoutUser(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request, res http.ResponseWriter) error {
	return m.Called(ctx, sessionCtxData, req, res).Error(0)
}