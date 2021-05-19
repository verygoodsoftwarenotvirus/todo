package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAccountStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.True(t, IsValidAccountStatus(string(GoodStandingAccountStatus)))
		assert.False(t, IsValidAccountStatus("blah"))
	})
}

func TestUser_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		actual := User{
			Username:        "old_username",
			HashedPassword:  "hashed_pass",
			TwoFactorSecret: "two factor secret",
		}
		exampleInput := User{
			Username:        "new_username",
			HashedPassword:  "updated_hashed_pass",
			TwoFactorSecret: "new fancy secret",
		}

		actual.Update(&exampleInput)
		assert.Equal(t, exampleInput, actual)
	})
}

func TestUser_IsBanned(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &User{ServiceAccountStatus: BannedUserAccountStatus}

		assert.True(t, x.IsBanned())
	})
}

func TestPasswordUpdateInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &PasswordUpdateInput{
			NewPassword:     t.Name(),
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx, 1))
	})
}

func TestTOTPSecretRefreshInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &TOTPSecretRefreshInput{
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestTOTPSecretVerificationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &TOTPSecretVerificationInput{
			UserID:    123,
			TOTPToken: "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestUserCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &UserRegistrationInput{
			Username: t.Name(),
			Password: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx, 1, 1))
	})
}

func TestUserLoginInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &UserLoginInput{
			Username:  t.Name(),
			Password:  t.Name(),
			TOTPToken: "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx, 1, 1))
	})
}
