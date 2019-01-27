package auth_test

import (
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

const (
	examplePassword           = "Pa$$w0rdPa$$w0rdPa$$w0rdPa$$w0rd"
	weakHashedExamplePassword = "$2a$10$iXvFoFHDudDpzhIhPbJIAukZ7QxkSVx6WmUfd01MG8zO5C0E8JCGC"
	hashedExamplePassword     = "$2a$13$hxMAo/ZRDmyaWcwvIem/vuUJkmeNytg3rwHUj6bRZR1d/cQHXjFvW"
	exampleTwoFactorSecret    = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func TestBcrypt_HashPassword(T *testing.T) {
	T.Parallel()

	x := auth.NewBcrypt(nil)

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		actual, err := x.HashPassword("password")
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}
func TestBcrypt_PasswordMatches(T *testing.T) {
	T.Parallel()

	x := auth.NewBcrypt(nil)

	T.Run("normal usage", func(t *testing.T) {
		t.Parallel()

		actual := x.PasswordMatches(hashedExamplePassword, examplePassword, nil)
		assert.True(t, actual)
	})

	T.Run("when passwords don't match", func(t *testing.T) {
		t.Parallel()

		actual := x.PasswordMatches(hashedExamplePassword, "password", nil)
		assert.False(t, actual)
	})
}

func TestBcrypt_PasswordIsAcceptable(T *testing.T) {
	T.Parallel()

	x := auth.NewBcrypt(nil)

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		assert.True(t, x.PasswordIsAcceptable(examplePassword))
		assert.False(t, x.PasswordIsAcceptable("hi there"))
	})
}
func TestBcrypt_ValidateLogin(T *testing.T) {
	T.Parallel()

	x := auth.NewBcrypt(nil)

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		code, err := totp.GenerateCode(exampleTwoFactorSecret, time.Now())
		assert.NoError(t, err, "error generating code to validate login")

		valid, err := x.ValidateLogin(
			hashedExamplePassword,
			examplePassword,
			exampleTwoFactorSecret,
			code,
		)
		assert.NoError(t, err, "unexpected eror encountered validating login: %v", err)
		assert.True(t, valid)
	})
}
