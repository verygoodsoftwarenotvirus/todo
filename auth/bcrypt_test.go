package auth_test

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"

	"github.com/opentracing/opentracing-go"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

const (
	examplePassword        = "Pa$$w0rdPa$$w0rdPa$$w0rdPa$$w0rd"
	hashedExamplePassword  = "$2a$13$hxMAo/ZRDmyaWcwvIem/vuUJkmeNytg3rwHUj6bRZR1d/cQHXjFvW"
	exampleTwoFactorSecret = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

var (
	tracer = &opentracing.NoopTracer{}
)

func TestBcrypt_HashPassword(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcrypt(auth.DefaultBcryptHashCost, nil, tracer)

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		tctx := context.Background()

		actual, err := x.HashPassword(tctx, "password")
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}
func TestBcrypt_PasswordMatches(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcrypt(auth.DefaultBcryptHashCost, nil, tracer)

	T.Run("normal usage", func(t *testing.T) {
		t.Parallel()
		tctx := context.Background()

		actual := x.PasswordMatches(tctx, hashedExamplePassword, examplePassword, nil)
		assert.True(t, actual)
	})

	T.Run("when passwords don't match", func(t *testing.T) {
		t.Parallel()
		tctx := context.Background()

		actual := x.PasswordMatches(tctx, hashedExamplePassword, "password", nil)
		assert.False(t, actual)
	})
}

func TestBcrypt_PasswordIsAcceptable(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcrypt(auth.DefaultBcryptHashCost, nil, tracer)

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		assert.True(t, x.PasswordIsAcceptable(examplePassword))
		assert.False(t, x.PasswordIsAcceptable("hi there"))
	})
}
func TestBcrypt_ValidateLogin(T *testing.T) {
	T.Parallel()

	x := auth.ProvideBcrypt(auth.DefaultBcryptHashCost, nil, tracer)

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		code, err := totp.GenerateCode(exampleTwoFactorSecret, time.Now())
		assert.NoError(t, err, "error generating code to validate login")

		valid, err := x.ValidateLogin(
			context.Background(),
			hashedExamplePassword,
			examplePassword,
			exampleTwoFactorSecret,
			code,
		)
		assert.NoError(t, err, "unexpected eror encountered validating login: %v", err)
		assert.True(t, valid)
	})
}

func TestPasswordIsAcceptable(T *testing.T) {
	T.SkipNow()

}

func TestValidateLogin(T *testing.T) {
	T.SkipNow()

}

func TestProvideBcrypt(T *testing.T) {
	T.SkipNow()

}

func TestHashPassword(T *testing.T) {
	T.SkipNow()

}

func TestPasswordMatches(T *testing.T) {
	T.SkipNow()

}
