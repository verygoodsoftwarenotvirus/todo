package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/stretchr/testify/mock"
)

var _ models.UserDataServer = (*UserDataServer)(nil)

// UserDataServer is a mocked models.UserDataServer for testing.
type UserDataServer struct {
	mock.Mock
}

// UserLoginInputMiddleware is a mock method to satisfy our interface requirements.
func (m *UserDataServer) UserLoginInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// UserInputMiddleware is a mock method to satisfy our interface requirements.
func (m *UserDataServer) UserInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// PasswordUpdateInputMiddleware is a mock method to satisfy our interface requirements.
func (m *UserDataServer) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// TOTPSecretVerificationInputMiddleware is a mock method to satisfy our interface requirements.
func (m *UserDataServer) TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// TOTPSecretRefreshInputMiddleware is a mock method to satisfy our interface requirements.
func (m *UserDataServer) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ListHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// SelfHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) SelfHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// TOTPSecretVerificationHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) TOTPSecretVerificationHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// NewTOTPSecretHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// UpdatePasswordHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) UpdatePasswordHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ArchiveHandler is a mock method to satisfy our interface requirements.
func (m *UserDataServer) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// AuditEntryHandler implements our interface requirements.
func (m *UserDataServer) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}
