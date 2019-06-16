package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.UserDataServer = (*UserDataServer)(nil)

// UserDataServer describes a structure capable of serving traffic related to users
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
func (m *UserDataServer) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// NewTOTPSecretHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// UpdatePasswordHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) UpdatePasswordHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// DeleteHandler is a mock method to satisfy our interface requirements
func (m *UserDataServer) DeleteHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ExportData is a mock method to satisfy our interface requirements
func (m *UserDataServer) ExportData(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}
