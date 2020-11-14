package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
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

func TestUser_ToSessionInfo(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleInput := User{
			ID:      12345,
			IsAdmin: true,
		}

		expected := &SessionInfo{
			UserID:      exampleInput.ID,
			UserIsAdmin: exampleInput.IsAdmin,
		}
		actual := exampleInput.ToSessionInfo()

		assert.Equal(t, expected, actual)
	})
}
