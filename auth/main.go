package auth

import (
	"crypto/rand"
	"errors"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	// ErrInvalidTwoFactorCode indicates that a provided two factor code is invalid
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
	// ErrPasswordHashTooWeak indicates that a provided password hash is too weak
	ErrPasswordHashTooWeak = errors.New("password's hash is too weak")
)

// PasswordHasher hashes passwords
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	PasswordIsAcceptable(password string) bool
	PasswordMatches(hashedPassword, providedPassword string, salt []byte) bool
}

// Enticator is a poorly named Authenticator interface
type Enticator interface {
	PasswordHasher
	ValidateLogin(hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error)
}

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}
