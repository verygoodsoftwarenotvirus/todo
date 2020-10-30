package auth

import (
	"context"
	"errors"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"

	"github.com/pquerna/otp/totp"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCostCompensation     = 2
	defaultMinimumPasswordSize = 16

	// DefaultBcryptHashCost is what it says on the tin.
	DefaultBcryptHashCost = BcryptHashCost(bcrypt.DefaultCost + bcryptCostCompensation)
)

var (
	_ Authenticator = (*BcryptAuthenticator)(nil)

	// ErrCostTooLow indicates that a password has too low a Bcrypt cost.
	ErrCostTooLow = errors.New("stored password's cost is too low")
)

type (
	// BcryptAuthenticator is our bcrypt-based authenticator
	BcryptAuthenticator struct {
		logger              logging.Logger
		hashCost            uint
		minimumPasswordSize uint
	}

	// BcryptHashCost is an arbitrary type alias for dependency injection's sake.
	BcryptHashCost uint
)

// ProvideBcryptAuthenticator returns a bcrypt powered Authenticator.
func ProvideBcryptAuthenticator(hashCost BcryptHashCost, logger logging.Logger) Authenticator {
	ba := &BcryptAuthenticator{
		logger:              logger.WithName("bcrypt"),
		hashCost:            uint(math.Min(float64(DefaultBcryptHashCost), float64(hashCost))),
		minimumPasswordSize: defaultMinimumPasswordSize,
	}
	return ba
}

// HashPassword takes a password and hashes it using bcrypt.
func (b *BcryptAuthenticator) HashPassword(ctx context.Context, password string) (string, error) {
	_, span := tracing.StartSpan(ctx, "HashPassword")
	defer span.End()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), int(b.hashCost))
	return string(hashedPass), err
}

// ValidateLogin validates a login attempt by:
// 1. checking that the provided password matches the stored hashed password
// 2. checking that the provided hashed password isn't too weak, and returning an error otherwise
// 3. checking that the temporary one-time password provided jives with the stored two factor secret
func (b *BcryptAuthenticator) ValidateLogin(
	ctx context.Context,
	hashedPassword,
	providedPassword,
	twoFactorSecret,
	twoFactorCode string,
	_ []byte,
) (passwordMatches bool, err error) {
	ctx, span := tracing.StartSpan(ctx, "ValidateLogin")
	defer span.End()

	passwordMatches = b.PasswordMatches(ctx, hashedPassword, providedPassword, nil)
	if !passwordMatches {
		return false, ErrPasswordDoesNotMatch
	}

	if err := b.hashedPasswordIsTooWeak(ctx, hashedPassword); err != nil {
		// NOTE: this can end up with a return set where passwordMatches is true and the err is not nil.
		// This is the valid case in the event the user has logged in with a valid password, but the
		// bcrypt cost has been raised since they last logged in.
		return passwordMatches, fmt.Errorf("validating password: %w", err)
	}

	if !totp.Validate(twoFactorCode, twoFactorSecret) {
		b.logger.WithValues(map[string]interface{}{
			"password_matches": passwordMatches,
			"2fa_secret":       twoFactorSecret,
			"provided_code":    twoFactorCode,
		}).Debug("invalid code provided")

		return passwordMatches, ErrInvalidTwoFactorCode
	}

	return passwordMatches, nil
}

// PasswordMatches validates whether or not a bcrypt-hashed password matches a provided password
func (b *BcryptAuthenticator) PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, _ []byte) bool {
	_, span := tracing.StartSpan(ctx, "PasswordMatches")
	defer span.End()

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword)) == nil
}

// hashedPasswordIsTooWeak determines if a given hashed password was hashed with too weak a bcrypt cost.
func (b *BcryptAuthenticator) hashedPasswordIsTooWeak(ctx context.Context, hashedPassword string) error {
	_, span := tracing.StartSpan(ctx, "hashedPasswordIsTooWeak")
	defer span.End()

	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		return fmt.Errorf("checking hashed password cost: %w", err)
	}

	if uint(cost) < b.hashCost {
		return ErrCostTooLow
	}

	return nil
}

// PasswordIsAcceptable takes a password and returns whether or not it satisfies the authenticator.
func (b *BcryptAuthenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}
