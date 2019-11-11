package mock

import (
	"net/http"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.UserDataServer = (*UserDataServer)(nil)

// UserDataServer is a mocked models.UserDataServer for testing
type UserDataServer struct {
	mock.Mock
}

// UserLoginInputMiddleware is a mock method to satisfy our interface requirements
func (m *UserDataServer) UserLoginInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// UserInputMiddleware is a mock method to satisfy our interface requirements
func (m *UserDataServer) UserInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// PasswordUpdateInputMiddleware is a mock method to satisfy our interface requirements
func (m *UserDataServer) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// TOTPSecretRefreshInputMiddleware is a mock method to satisfy our interface requirements
func (m *UserDataServer) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ListHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) ListHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// CreateHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) CreateHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// ReadHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) ReadHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// NewTOTPSecretHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) NewTOTPSecretHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// UpdatePasswordHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) UpdatePasswordHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// ArchiveHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) ArchiveHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}
