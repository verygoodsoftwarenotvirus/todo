package auth

import (
	"errors"
)

var (
	ErrCostTooLow           = errors.New("stored password's cost is too low")
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
	ErrPasswordHashTooWeak  = errors.New("password's hash is too weak")
)

type Enticator interface {
	HashPassword(password string) (string, error)
	PasswordIsAcceptable(password string) bool
	PasswordMatches(hashedPassword, providedPassword string) bool
	ValidateLogin(hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error)
}
