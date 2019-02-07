package auth

import (
	"context"
	"errors"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/opentracing/opentracing-go"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptHashCost is what it says on the tin
	DefaultBcryptHashCost      = BcryptHashCost(bcrypt.DefaultCost + 2)
	defaultMinimumPasswordSize = 16
)

var (
	_ Enticator = (*BcryptAuthenticator)(nil)

	// ErrCostTooLow indicates that a password has too low a Bcrypt cost
	ErrCostTooLow = errors.New("stored password's cost is too low")
)

// BcryptAuthenticator is our bcrypt-based authenticator
type BcryptAuthenticator struct {
	logger              logging.Logger
	hashCost            uint
	minimumPasswordSize uint
	tracer              opentracing.Tracer
}

// BcryptHashCost is an arbitrary type alias for dependency injection's sake.
type BcryptHashCost uint

// ProvideBcrypt returns a Bcrypt-powered Enticator
func ProvideBcrypt(hashCost BcryptHashCost, logger logging.Logger, tracer Tracer) Enticator {
	ba := &BcryptAuthenticator{
		logger:              logger,
		tracer:              tracer,
		hashCost:            uint(math.Min(float64(DefaultBcryptHashCost), float64(hashCost))),
		minimumPasswordSize: defaultMinimumPasswordSize,
	}
	return ba
}

// HashPassword takes a password and hashes it using bcrypt
func (b *BcryptAuthenticator) HashPassword(ctx context.Context, password string) (string, error) {
	span := tracing.FetchSpanFromContext(ctx, b.tracer, "HashPassword")
	defer span.Finish()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), int(b.hashCost))
	return string(hashedPass), err
}

// PasswordMatches validates whether or not a bcrypt-hashed password matches a provided password
func (b *BcryptAuthenticator) PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, _ []byte) bool {
	span := tracing.FetchSpanFromContext(ctx, b.tracer, "PasswordMatches")
	defer span.Finish()

	matches := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(providedPassword),
	) == nil
	tooWeak := b.hashedPasswordIsTooWeak(hashedPassword)

	b.logger.WithValues(map[string]interface{}{
		"too_weak": tooWeak,
		"matches":  matches,
	}).Debug("evaluated password match")

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
func (b *BcryptAuthenticator) ValidateLogin(ctx context.Context, hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error) {
	span := tracing.FetchSpanFromContext(ctx, b.tracer, "ValidateLogin")
	defer span.Finish()

	passwordMatches := b.PasswordMatches(ctx, hashedPassword, providedPassword, nil)
	if !totp.Validate(twoFactorCode, twoFactorSecret) {
		return passwordMatches, ErrInvalidTwoFactorCode
	}
	return passwordMatches, nil
}
