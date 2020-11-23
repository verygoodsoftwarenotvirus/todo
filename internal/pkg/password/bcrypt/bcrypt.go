package bcrypt

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	"github.com/pquerna/otp/totp"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

const (
	defaultMinimumPasswordSize = 16
	bcryptCostCompensation     = 2

	// DefaultHashCost is what it says on the tin.
	DefaultHashCost = HashCost(bcrypt.DefaultCost + bcryptCostCompensation)
)

var (
	_ password.Authenticator = (*Authenticator)(nil)

	// ErrCostTooLow indicates that a password has too low a Bcrypt cost.
	ErrCostTooLow = errors.New("stored password's cost is too low")
)

type (
	// Authenticator is our bcrypt-based authenticator.
	Authenticator struct {
		logger              logging.Logger
		hashCost            uint
		minimumPasswordSize uint
	}

	// HashCost is an arbitrary type alias for dependency injection's sake.
	HashCost uint
)

// ProvideAuthenticator returns a bcrypt powered Authenticator.
func ProvideAuthenticator(hashCost HashCost, logger logging.Logger) password.Authenticator {
	ba := &Authenticator{
		logger:              logger.WithName("bcrypt"),
		hashCost:            uint(math.Min(float64(DefaultHashCost), float64(hashCost))),
		minimumPasswordSize: defaultMinimumPasswordSize,
	}

	return ba
}

// HashPassword takes a password and hashes it using bcrypt.
func (b *Authenticator) HashPassword(ctx context.Context, passwordToHash string) (string, error) {
	_, span := tracing.StartSpan(ctx)
	defer span.End()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(passwordToHash), int(b.hashCost))

	return string(hashedPass), err
}

// ValidateLogin validates a login attempt by:
// 1. checking that the provided password matches the stored hashed password
// 2. checking that the provided hashed password isn't too weak, and returning an error otherwise
// 3. checking that the temporary one-time password provided jives with the stored two factor secret.
func (b *Authenticator) ValidateLogin(
	ctx context.Context,
	hashedPassword,
	providedPassword,
	twoFactorSecret,
	twoFactorCode string,
	_ []byte,
) (passwordMatches bool, err error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	passwordMatches = b.PasswordMatches(ctx, hashedPassword, providedPassword, nil)
	if !passwordMatches {
		return false, password.ErrPasswordDoesNotMatch
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

		return passwordMatches, password.ErrInvalidTwoFactorCode
	}

	return passwordMatches, nil
}

// PasswordMatches validates whether or not a bcrypt-hashed password matches a provided password.
func (b *Authenticator) PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, _ []byte) bool {
	_, span := tracing.StartSpan(ctx)
	defer span.End()

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword)) == nil
}

// hashedPasswordIsTooWeak determines if a given hashed password was hashed with too weak a bcrypt cost.
func (b *Authenticator) hashedPasswordIsTooWeak(ctx context.Context, hashedPassword string) error {
	_, span := tracing.StartSpan(ctx)
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
func (b *Authenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}
