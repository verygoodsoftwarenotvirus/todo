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
	ErrCostTooLow           = errors.New("stored password's cost is too low")
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
	ErrPasswordHashTooWeak  = errors.New("password's hash is too weak")
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	PasswordIsAcceptable(password string) bool
	PasswordMatches(hashedPassword, providedPassword string, salt []byte) bool
}

type Enticator2 interface {
	PasswordMatches(providedPassword string) bool
	ValidateLogin(providedPassword, twoFactorCode string) (bool, error)
}

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

// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func RandString(len uint64) (string, error) {
	b := make([]byte, uint64(math.Max(64, float64(len))))
	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}
