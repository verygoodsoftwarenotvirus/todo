package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
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

		actual.Update(&exampleInput)

		assert.Equal(t, exampleInput, actual)
	})
}
