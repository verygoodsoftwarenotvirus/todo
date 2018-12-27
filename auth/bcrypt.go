package auth

import (
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultHashCost            = uint(bcrypt.DefaultCost) + 3
	defaultMinimumPasswordSize = 16
)

var _ Enticator = (*BcryptAuthenticator)(nil)

type BcryptAuthenticator struct {
	logger              *logrus.Logger
	hashCost            uint
	minimumPasswordSize uint
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

func (b *BcryptAuthenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}

func (b *BcryptAuthenticator) ValidateLogin(hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error) {
	passwordMatches := b.PasswordMatches(hashedPassword, providedPassword, nil)
	if !totp.Validate(twoFactorCode, twoFactorSecret) {
		return passwordMatches, ErrInvalidTwoFactorCode
	}
	return passwordMatches, nil
}
