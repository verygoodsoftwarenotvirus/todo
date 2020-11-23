package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"

	"github.com/stretchr/testify/mock"
)

var _ password.Authenticator = (*Authenticator)(nil)

// Authenticator is a mock Authenticator.
type Authenticator struct {
	mock.Mock
}

// ValidateLogin satisfies our authenticator interface.
func (m *Authenticator) ValidateLogin(
	ctx context.Context,
	hashedPassword,
	providedPassword,
	twoFactorSecret,
	twoFactorCode string,
	salt []byte,
) (valid bool, err error) {
	args := m.Called(
		ctx,
		hashedPassword,
		providedPassword,
		twoFactorSecret,
		twoFactorCode,
		salt,
	)

	return args.Bool(0), args.Error(1)
}

// PasswordIsAcceptable satisfies our authenticator interface.
func (m *Authenticator) PasswordIsAcceptable(pass string) bool {
	return m.Called(pass).Bool(0)
}

// HashPassword satisfies our authenticator interface.
func (m *Authenticator) HashPassword(ctx context.Context, pass string) (string, error) {
	args := m.Called(ctx, pass)
	return args.String(0), args.Error(1)
}

// PasswordMatches satisfies our authenticator interface.
func (m *Authenticator) PasswordMatches(
	ctx context.Context,
	hashedPassword,
	providedPassword string,
	salt []byte,
) bool {
	args := m.Called(
		ctx,
		hashedPassword,
		providedPassword,
		salt,
	)

	return args.Bool(0)
}
