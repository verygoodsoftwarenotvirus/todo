package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_Update(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		actual := User{
			Username:        "username",
			HashedPassword:  "hashed_pass",
			TwoFactorSecret: "two factor secret",
		}

		exampleInput := User{
			Username:        "newUsername",
			HashedPassword:  "updated_hashed_pass",
			TwoFactorSecret: "new fancy secret",
		}

		actual.Update(exampleInput)

		assert.Equal(t, exampleInput, actual)
	})
}
