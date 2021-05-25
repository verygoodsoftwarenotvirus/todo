package authentication

import (
	"context"
	"crypto/rand"
	"errors"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

var (
	// ErrInvalidTOTPToken indicates that a provided two factor code is invalid.
	ErrInvalidTOTPToken = errors.New("invalid two factor code")
	// ErrPasswordDoesNotMatch indicates that a provided passwords does not match.
	ErrPasswordDoesNotMatch = errors.New("password does not match")
)

type (
	// Hasher hashes passwords.
	Hasher interface {
		HashPassword(ctx context.Context, password string) (string, error)
	}

	// Authenticator authenticates users.
	Authenticator interface {
		Hasher

		ValidateLogin(ctx context.Context, hash, password, totpSecret, totpCode string) (bool, error)
	}
)
