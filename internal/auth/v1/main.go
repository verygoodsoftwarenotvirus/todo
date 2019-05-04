package auth

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/google/wire"
)

var (
	// ErrInvalidTwoFactorCode indicates that a provided two factor code is invalid
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
	// ErrPasswordHashTooWeak indicates that a provided password hash is too weak
	ErrPasswordHashTooWeak = errors.New("password's hash is too weak")

	// Providers represents what this package offers to external libraries in the way of consntructors
	Providers = wire.NewSet(
		ProvideBcrypt,
		ProvideBcryptHashCost,
	)
)

// ProvideBcryptHashCost provides a BcryptHashCost
func ProvideBcryptHashCost() BcryptHashCost {
	return DefaultBcryptHashCost
}

// PasswordHasher hashes passwords
type PasswordHasher interface {
	PasswordIsAcceptable(password string) bool
	HashPassword(ctx context.Context, password string) (string, error)
	PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, salt []byte) bool
}

// Authenticator is a poorly named Authenticator interface
type Authenticator interface {
	PasswordHasher

	ValidateLogin(
		ctx context.Context,
		HashedPassword string,
		Salt []byte,
		ProvidedPassword string,
		TwoFactorSecret string,
		TwoFactorCode string,
	) (valid bool, err error)
}

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}
