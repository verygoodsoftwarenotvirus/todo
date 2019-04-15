package auth

import (
	"context"
	"crypto/rand"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/google/wire"
	"github.com/opentracing/opentracing-go"
)

var (
	// ErrInvalidTwoFactorCode indicates that a provided two factor code is invalid
	ErrInvalidTwoFactorCode = errors.New("invalid two factor code")
	// ErrPasswordHashTooWeak indicates that a provided password hash is too weak
	ErrPasswordHashTooWeak = errors.New("password's hash is too weak")

	// Providers represents what this package offers to external libraries in the way of consntructors
	Providers = wire.NewSet(
		ProvideBcrypt,
		ProvideTracer,
		ProvideBcryptHashCost,
	)
)

// ProvideBcryptHashCost provides a BcryptHashCost
func ProvideBcryptHashCost() BcryptHashCost {
	return DefaultBcryptHashCost
}

// PasswordHasher hashes passwords
type PasswordHasher interface {
	PasswordIsAcceptable(password string) bool
	HashPassword(ctx context.Context, password string) (string, error)
	PasswordMatches(ctx context.Context, hashedPassword, providedPassword string, salt []byte) bool
}

// Tracer is an obligatory type alias we have for dependency injection's sake
type Tracer opentracing.Tracer

// ProvideTracer provides a Tracer
func ProvideTracer() Tracer {
	return tracing.ProvideTracer("password-authentication")
}

// Enticator is a poorly named Authenticator interface
type Enticator interface {
	PasswordHasher

	ValidateLogin(
		ctx context.Context,
		HashedPassword,
		ProvidedPassword,
		TwoFactorSecret,
		TwoFactorCode string,
	) (bool, error)
}

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}
