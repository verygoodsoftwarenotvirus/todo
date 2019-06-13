package auth

import (
	"context"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptHashCost is what it says on the tin
	DefaultBcryptHashCost      = BcryptHashCost(bcrypt.DefaultCost + 2)
	defaultMinimumPasswordSize = 16
)

var (
	_ Authenticator = (*BcryptAuthenticator)(nil)

	// ErrCostTooLow indicates that a password has too low a Bcrypt cost
	ErrCostTooLow = errors.New("stored password's cost is too low")
)

// BcryptAuthenticator is our bcrypt-based authenticator
type BcryptAuthenticator struct {
	logger              logging.Logger
	hashCost            uint
	minimumPasswordSize uint
}

// BcryptHashCost is an arbitrary type alias for dependency injection's sake.
type BcryptHashCost uint

// ProvideBcrypt returns a Bcrypt-powered Authenticator
func ProvideBcrypt(hashCost BcryptHashCost, logger logging.Logger) Authenticator {
	ba := &BcryptAuthenticator{
		logger:              logger.WithName("bcrypt"),
		hashCost:            uint(math.Min(float64(DefaultBcryptHashCost), float64(hashCost))),
		minimumPasswordSize: defaultMinimumPasswordSize,
	}
	return ba
}

// HashPassword takes a password and hashes it using bcrypt
func (b *BcryptAuthenticator) HashPassword(c context.Context, password string) (string, error) {
	_, span := trace.StartSpan(c, "HashPassword")
	defer span.End()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), int(b.hashCost))
	return string(hashedPass), err
}

// ValidateLogin validates a password and two factor code
func (b *BcryptAuthenticator) ValidateLogin(
	ctx context.Context,
	hashedPassword,
	providedPassword,
	twoFactorSecret,
	twoFactorCode string,
	salt []byte,
) (passwordMatches bool, err error) {
	ctx, span := trace.StartSpan(ctx, "ValidateLogin")
	defer span.End()

	passwordMatches = b.PasswordMatches(ctx, hashedPassword, providedPassword, nil)
	tooWeak := b.hashedPasswordIsTooWeak(hashedPassword)

	if !totp.Validate(twoFactorCode, twoFactorSecret) {
		b.logger.WithValues(map[string]interface{}{
			"password_matches": passwordMatches,
			"2fa_secret":       twoFactorSecret,
			"provided_code":    twoFactorCode,
		}).Debug("invalid code provided")

		return passwordMatches, ErrInvalidTwoFactorCode
	}

	if tooWeak {
		return passwordMatches, ErrCostTooLow
	}

	return passwordMatches, nil
}

// PasswordMatches validates whether or not a bcrypt-hashed password matches a provided password
func (b *BcryptAuthenticator) PasswordMatches(
	ctx context.Context,
	hashedPassword,
	providedPassword string,
	_ []byte) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword)) == nil
}

func (b *BcryptAuthenticator) hashedPasswordIsTooWeak(hashedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashedPassword))

	if err != nil || uint(cost) < b.hashCost {
		return true
	}

	return false
}

// PasswordIsAcceptable takes a password and returns whether or not it satisfies the authenticator
func (b *BcryptAuthenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}
