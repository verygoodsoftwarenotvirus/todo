package auth_test

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

const (
	examplePassword             = "Pa$$w0rdPa$$w0rdPa$$w0rdPa$$w0rd"
	weaklyHashedExamplePassword = "$2a$04$7G7dHZe7MeWjOMsYKO8uCu/CRKnDMMBHOfXaB6YgyQL/cl8nhwf/2"
	hashedExamplePassword       = "$2a$13$hxMAo/ZRDmyaWcwvIem/vuUJkmeNytg3rwHUj6bRZR1d/cQHXjFvW"
	exampleTwoFactorSecret      = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func TestBcrypt_HashPassword(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcryptAuthenticator(auth.DefaultBcryptHashCost, noop.NewLogger())

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		actual, err := x.HashPassword(ctx, "password")
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestBcrypt_PasswordMatches(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcryptAuthenticator(auth.DefaultBcryptHashCost, noop.NewLogger())

	T.Run("normal usage", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		actual := x.PasswordMatches(ctx, hashedExamplePassword, examplePassword, nil)
		assert.True(t, actual)
	})

	T.Run("when passwords don't match", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		actual := x.PasswordMatches(ctx, hashedExamplePassword, "password", nil)
		assert.False(t, actual)
	})
}

func TestBcrypt_PasswordIsAcceptable(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcryptAuthenticator(auth.DefaultBcryptHashCost, noop.NewLogger())

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		assert.True(t, x.PasswordIsAcceptable(examplePassword))
		assert.False(t, x.PasswordIsAcceptable("hi there"))
	})
}

func TestBcrypt_ValidateLogin(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcryptAuthenticator(auth.DefaultBcryptHashCost, noop.NewLogger())

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		code, err := totp.GenerateCode(exampleTwoFactorSecret, time.Now().UTC())
		assert.NoError(t, err, "error generating code to validate login")

		valid, err := x.ValidateLogin(
			ctx,
			hashedExamplePassword,
			examplePassword,
			exampleTwoFactorSecret,
			code,
			nil,
		)
		assert.NoError(t, err, "unexpected error encountered validating login: %v", err)
		assert.True(t, valid)
	})

	T.Run("with weak hash", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		code, err := totp.GenerateCode(exampleTwoFactorSecret, time.Now().UTC())
		assert.NoError(t, err, "error generating code to validate login")

		valid, err := x.ValidateLogin(
			ctx,
			weaklyHashedExamplePassword,
			examplePassword,
			exampleTwoFactorSecret,
			code,
			nil,
		)
		assert.Error(t, err, "unexpected error encountered validating login: %v", err)
		assert.True(t, valid)
	})

	T.Run("with non-matching password", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		code, err := totp.GenerateCode(exampleTwoFactorSecret, time.Now().UTC())
		assert.NoError(t, err, "error generating code to validate login")

		valid, err := x.ValidateLogin(
			ctx,
			hashedExamplePassword,
			"examplePassword",
			exampleTwoFactorSecret,
			code,
			nil,
		)
		assert.Error(t, err, "unexpected error encountered validating login: %v", err)
		assert.Equal(t, err, auth.ErrPasswordDoesNotMatch)
		assert.False(t, valid)
	})

	T.Run("with invalid code", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		valid, err := x.ValidateLogin(
			ctx,
			hashedExamplePassword,
			examplePassword,
			exampleTwoFactorSecret,
			"CODE",
			nil,
		)
		assert.Error(t, err, "unexpected error encountered validating login: %v", err)
		assert.True(t, valid)
	})
}

func TestProvideBcrypt(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		auth.ProvideBcryptAuthenticator(auth.DefaultBcryptHashCost, noop.NewLogger())
	})
}
