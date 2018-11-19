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
	ErrCostTooLow           = errors.New("stored password's cost is too low")
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
)

var _ Enticator = (*BcryptAuthenticator)(nil)

type BcryptAuthenticator struct {
	logger              *logrus.Logger
	hashCost            uint
	minimumPasswordSize uint
}

func New(hashCost, minimumPasswordSize uint) *BcryptAuthenticator {
	return &BcryptAuthenticator{
		hashCost:            hashCost,
		minimumPasswordSize: minimumPasswordSize,
	}
}

func NewBcrypt(logger *logrus.Logger) *BcryptAuthenticator {
	if logger == nil {
		logger = logrus.New()
	}
	return &BcryptAuthenticator{
		logger:              logger,
		hashCost:            defaultHashCost,
		minimumPasswordSize: defaultMinimumPasswordSize,
	}
}

func (b *BcryptAuthenticator) HashPassword(password string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), int(b.hashCost))
	return string(hashedPass), err
}

func (b *BcryptAuthenticator) PasswordMatches(hashedPassword, providedPassword string) bool {
	matches := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword)) == nil
	tooWeak := b.passwordIsTooWeak([]byte(hashedPassword))

	return matches && !tooWeak
}

func (b *BcryptAuthenticator) passwordIsTooWeak(hashedPassword []byte) bool {
	cost, err := bcrypt.Cost(hashedPassword)

	if err != nil || uint(cost) != b.hashCost {
		return true
	}

	return false
}

func (b *BcryptAuthenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}

func (b *BcryptAuthenticator) ValidateLogin(hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error) {
	passwordMatches := b.PasswordMatches(hashedPassword, providedPassword)
	tokenIsValid := totp.Validate(twoFactorCode, twoFactorSecret)
	if !tokenIsValid {
		return passwordMatches, ErrInvalidTwoFactorCode
	}

	return passwordMatches, nil
}
