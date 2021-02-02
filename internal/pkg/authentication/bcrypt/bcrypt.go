package bcrypt

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	observableName = "bcrypt"
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
	_ authentication.Authenticator = (*Authenticator)(nil)

	// ErrCostTooLow indicates that a authentication has too low a Bcrypt cost.
	ErrCostTooLow = errors.New("stored authentication's cost is too low")
)

type (
	// Authenticator is our bcrypt-based authenticator.
	Authenticator struct {
		logger              logging.Logger
		hashCost            uint
		minimumPasswordSize uint
		tracer              tracing.Tracer
	}

	// HashCost is a type alias for dependency injection's sake.
	HashCost uint
)

// ProvideAuthenticator returns a bcrypt powered Authenticator.
func ProvideAuthenticator(hashCost HashCost, logger logging.Logger) authentication.Authenticator {
	ba := &Authenticator{
		logger:              logger.WithName(observableName),
		hashCost:            uint(math.Min(float64(DefaultHashCost), float64(hashCost))),
		minimumPasswordSize: defaultMinimumPasswordSize,
		tracer:              tracing.NewTracer(observableName),
	}

	return ba
}

// HashPassword takes a authentication and hashes it using bcrypt.
func (b *Authenticator) HashPassword(ctx context.Context, passwordToHash string) (string, error) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(passwordToHash), int(b.hashCost))

	return string(hashedPass), err
}

// ValidateLogin validates a login attempt by:
// 1. checking that the provided authentication matches the stored hashed authentication
// 2. checking that the provided hashed authentication isn't too weak, and returning an error otherwise
// 3. checking that the temporary one-time authentication provided jives with the stored two factor secret.
func (b *Authenticator) ValidateLogin(
	ctx context.Context,
	hashedPassword,
	providedPassword,
	twoFactorSecret,
	twoFactorCode string,
	_ []byte,
) (passwordMatches bool, err error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	passwordMatches = b.PasswordMatches(ctx, hashedPassword, providedPassword, nil)
	if !passwordMatches {
		return false, authentication.ErrPasswordDoesNotMatch
	}

	if err := b.hashedPasswordIsTooWeak(ctx, hashedPassword); err != nil {
		// NOTE: this can end up with a return set where passwordMatches is true and the err is not nil.
		// This is the valid case in the event the user has logged in with a valid authentication, but the
		// bcrypt cost has been raised since they last logged in.
		return passwordMatches, fmt.Errorf("validating authentication: %w", err)
	}

	if !totp.Validate(twoFactorCode, twoFactorSecret) {
		b.logger.WithValues(map[string]interface{}{
			"password_matches": passwordMatches,
			"2fa_secret":       twoFactorSecret,
			"provided_code":    twoFactorCode,
		}).Debug("invalid code provided")

		return passwordMatches, authentication.ErrInvalidTwoFactorCode
	}

	return passwordMatches, nil
}

// PasswordMatches validates whether or not a bcrypt-hashed authentication matches a provided authentication.
func (b *Authenticator) PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, _ []byte) bool {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword)) == nil
}

// hashedPasswordIsTooWeak determines if a given hashed authentication was hashed with too weak a bcrypt cost.
func (b *Authenticator) hashedPasswordIsTooWeak(ctx context.Context, hashedPassword string) error {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		return fmt.Errorf("checking hashed authentication cost: %w", err)
	}

	if uint(cost) < b.hashCost {
		return ErrCostTooLow
	}

	return nil
}

// PasswordIsAcceptable takes a authentication and returns whether or not it satisfies the authenticator.
func (b *Authenticator) PasswordIsAcceptable(pass string) bool {
	return uint(len(pass)) >= b.minimumPasswordSize
}
