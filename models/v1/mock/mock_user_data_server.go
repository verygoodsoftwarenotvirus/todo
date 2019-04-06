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

// List is a mock method to satisfy our interface requirements
func (m *UserDataServer) List(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// Create is a mock method to satisfy our interface requirements
func (m *UserDataServer) Create(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// Read is a mock method to satisfy our interface requirements
func (m *UserDataServer) Read(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// NewTOTPSecret is a mock method to satisfy our interface requirements
func (m *UserDataServer) NewTOTPSecret(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// UpdatePassword is a mock method to satisfy our interface requirements
func (m *UserDataServer) UpdatePassword(res http.ResponseWriter, req *http.Request) {
	m.Called()
}

// Delete is a mock method to satisfy our interface requirements
func (m *UserDataServer) Delete(res http.ResponseWriter, req *http.Request) {
	m.Called()
}
