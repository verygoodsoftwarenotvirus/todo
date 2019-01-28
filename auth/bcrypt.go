package auth

import (
	"errors"

	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultHashCost            = uint(bcrypt.DefaultCost) + 3
	defaultMinimumPasswordSize = 16
)

var (
	_ Enticator = (*BcryptAuthenticator)(nil)

	// ErrCostTooLow indicates that a password has too low a Bcrypt cost
	ErrCostTooLow = errors.New("stored password's cost is too low")
)

// BcryptAuthenticator is our bcrypt-based authenticator
type BcryptAuthenticator struct {
	logger              *logrus.Logger
	hashCost            uint
	minimumPasswordSize uint
}

// NewBcrypt returns a Bcrypt-powered Enticator
func NewBcrypt(logger *logrus.Logger) Enticator {
	if logger == nil {
		logger = logrus.New()
	}
	return &BcryptAuthenticator{
		logger:              logger,
		hashCost:            defaultHashCost,
		minimumPasswordSize: defaultMinimumPasswordSize,
	}
}

// HashPassword takes a password and hashes it using bcrypt
func (b *BcryptAuthenticator) HashPassword(password string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), int(b.hashCost))
	return string(hashedPass), err
}

// PasswordMatches validates whether or not a bcrypt-hashed password matches a provided password
func (b *BcryptAuthenticator) PasswordMatches(hashedPassword, providedPassword string, salt []byte) bool {
	matches := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(providedPassword),
	) == nil
	tooWeak := b.hashedPasswordIsTooWeak(hashedPassword)

	return matches && !tooWeak
}

func (b *BcryptAuthenticator) hashedPasswordIsTooWeak(hashedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashedPassword))

	if err != nil || uint(cost) != b.hashCost {
		return true
	}

	return false
}

// PasswordIsAcceptable takes a password and returns whether or not it satisfies the authenticator
func (b *BcryptAuthenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}

// ValidateLogin validates a password and two factor code
func (b *BcryptAuthenticator) ValidateLogin(hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error) {
	passwordMatches := b.PasswordMatches(hashedPassword, providedPassword, nil)
	if !totp.Validate(twoFactorCode, twoFactorSecret) {
		return passwordMatches, ErrInvalidTwoFactorCode
	}
	return passwordMatches, nil
}
