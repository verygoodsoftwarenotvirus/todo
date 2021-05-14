package mock

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

// UsersService is a mock types.UsersService.
type UsersService struct {
	mock.Mock
}

// UserRegistrationInputMiddleware satisfies our interface contract.
func (m *UsersService) UserRegistrationInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// PasswordUpdateInputMiddleware satisfies our interface contract.
func (m *UsersService) PasswordUpdateInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// TOTPSecretRefreshInputMiddleware satisfies our interface contract.
func (m *UsersService) TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// TOTPSecretVerificationInputMiddleware satisfies our interface contract.
func (m *UsersService) TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

// ListHandler satisfies our interface contract.
func (m *UsersService) ListHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// AuditEntryHandler satisfies our interface contract.
func (m *UsersService) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// CreateHandler satisfies our interface contract.
func (m *UsersService) CreateHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ReadHandler satisfies our interface contract.
func (m *UsersService) ReadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// SelfHandler satisfies our interface contract.
func (m *UsersService) SelfHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// UsernameSearchHandler satisfies our interface contract.
func (m *UsersService) UsernameSearchHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// NewTOTPSecretHandler satisfies our interface contract.
func (m *UsersService) NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// TOTPSecretVerificationHandler satisfies our interface contract.
func (m *UsersService) TOTPSecretVerificationHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// UpdatePasswordHandler satisfies our interface contract.
func (m *UsersService) UpdatePasswordHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// AvatarUploadHandler satisfies our interface contract.
func (m *UsersService) AvatarUploadHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// ArchiveHandler satisfies our interface contract.
func (m *UsersService) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// RegisterUser satisfies our interface contract.
func (m *UsersService) RegisterUser(ctx context.Context, registrationInput *types.UserRegistrationInput) (*types.UserCreationResponse, error) {
	returnValues := m.Called(ctx, registrationInput)

	return returnValues.Get(0).(*types.UserCreationResponse), returnValues.Error(1)
}

// VerifyUserTwoFactorSecret satisfies our interface contract.
func (m *UsersService) VerifyUserTwoFactorSecret(ctx context.Context, input *types.TOTPSecretVerificationInput) error {
	return m.Called(ctx, input).Error(0)
}
