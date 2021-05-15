package passwords

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/alexedwards/argon2id"
	"github.com/pquerna/otp/totp"
)

const (
	argon2IterationCount = 1
	argon2ThreadCount    = 2
	argon2SaltLength     = 16
	argon2KeyLength      = 32
	sixtyFourMegabytes   = 64 * 1024
)

var argonParams = &argon2id.Params{
	Memory:      sixtyFourMegabytes,
	Iterations:  argon2IterationCount,
	Parallelism: argon2ThreadCount,
	SaltLength:  argon2SaltLength,
	KeyLength:   argon2KeyLength,
}

type (
	// Argon2Authenticator is our argon2-based authenticator.
	Argon2Authenticator struct {
		logger logging.Logger
		tracer tracing.Tracer
	}
)

// ProvideArgon2Authenticator returns an argon2 powered Argon2Authenticator.
func ProvideArgon2Authenticator(logger logging.Logger) Authenticator {
	ba := &Argon2Authenticator{
		logger: logging.EnsureLogger(logger).WithName(argon2Provider),
		tracer: tracing.NewTracer(argon2Provider),
	}

	return ba
}

// HashPassword takes a password and hashes it using argon2.
func (a *Argon2Authenticator) HashPassword(ctx context.Context, password string) (string, error) {
	_, span := a.tracer.StartSpan(ctx)
	defer span.End()

	return argon2id.CreateHash(password, argonParams)
}

// ValidateLogin validates a login attempt by:
//  - checking that the provided authentication matches the provided hashed passwords.
//  - checking that the temporary one-time authentication provided jives with the provided two factor secret.
func (a *Argon2Authenticator) ValidateLogin(ctx context.Context, hash, password, totpSecret, totpCode string) (bool, error) {
	_, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger

	passwordMatches, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, observability.PrepareError(err, a.logger, span, "comparing argon2 hashed password")
	} else if !passwordMatches {
		return false, ErrPasswordDoesNotMatch
	}

	if !totp.Validate(totpCode, totpSecret) {
		logger.WithValues(map[string]interface{}{
			"password_matches": passwordMatches,
			"provided_code":    totpCode,
		}).Debug("invalid code provided")

		return passwordMatches, ErrInvalidTOTPToken
	}

	return passwordMatches, nil
}
