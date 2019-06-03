package mock

import (
	"context"

	libauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"

	"github.com/stretchr/testify/mock"
)

var _ libauth.Authenticator = (*Authenticator)(nil)

// Authenticator is a mock Authenticator
type Authenticator struct {
	mock.Mock
}

// ValidateLogin satisfies our authenticator interface
func (m *Authenticator) ValidateLogin(
	ctx context.Context,
	HashedPassword string,
	Salt []byte,
	ProvidedPassword string,
	TwoFactorSecret string,
	TwoFactorCode string,
) (valid bool, err error) {
	args := m.Called(ctx,
		HashedPassword,
		Salt,
		ProvidedPassword,
		TwoFactorSecret,
		TwoFactorCode,
	)
	return args.Bool(0), args.Error(1)
}

// PasswordIsAcceptable satisfies our authenticator interface
func (m *Authenticator) PasswordIsAcceptable(password string) bool {
	return m.Called(password).Bool(0)
}

// HashPassword satisfies our authenticator interface
func (m *Authenticator) HashPassword(ctx context.Context, password string) (string, error) {
	args := m.Called(ctx, password)
	return args.String(0), args.Error(1)
}

// PasswordMatches satisfies our authenticator interface
func (m *Authenticator) PasswordMatches(
	ctx context.Context,
	hashedPassword,
	providedPassword string,
	salt []byte,
) bool {
	args := m.Called(ctx,
		hashedPassword,
		providedPassword,
		salt,
	)
	return args.Bool(0)
}
