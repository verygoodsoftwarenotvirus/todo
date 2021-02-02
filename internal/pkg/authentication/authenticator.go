package authentication

import (
	"context"
	"errors"
)

var (
	// ErrInvalidTwoFactorCode indicates that a provided two factor code is invalid.
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
	// ErrPasswordHashTooWeak indicates that a provided authentication hash is too weak.
	ErrPasswordHashTooWeak = errors.New("authentication's hash is too weak")
	// ErrPasswordDoesNotMatch indicates that a provided authentication does not match.
	ErrPasswordDoesNotMatch = errors.New("authentication's hash is too weak")
)

type (
	// Hasher hashes passwords.
	Hasher interface {
		PasswordIsAcceptable(password string) bool
		HashPassword(ctx context.Context, password string) (string, error)
		PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, salt []byte) bool
	}

	// Authenticator is a poorly named Authenticator interface.
	Authenticator interface {
		Hasher

		ValidateLogin(
			ctx context.Context,
			hashedPassword,
			providedPassword,
			twoFactorSecret,
			twoFactorCode string,
			salt []byte,
		) (valid bool, err error)
	}
)
