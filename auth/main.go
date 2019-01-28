package auth

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"math"
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

// RandString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func RandString(len uint64) (string, error) {
	b := make([]byte, uint64(math.Max(64, float64(len))))
	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}
