package authentication

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Authenticator = (*MockAuthenticator)(nil)

// MockAuthenticator is a mock Authenticator.
type MockAuthenticator struct {
	mock.Mock
}

// ValidateLogin satisfies our authenticator interface.
func (m *MockAuthenticator) ValidateLogin(ctx context.Context, hash, password, totpSecret, totpCode string) (bool, error) {
	args := m.Called(ctx, hash, password, totpSecret, totpCode)

	return args.Bool(0), args.Error(1)
}

// HashPassword satisfies our authenticator interface.
func (m *MockAuthenticator) HashPassword(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}
