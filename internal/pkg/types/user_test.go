package types

import (
	"encoding/json"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_JSONUnmarshal(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		exampleInput := User{
			Username:         "new_username",
			HashedPassword:   "updated_hashed_pass",
			TwoFactorSecret:  "new fancy secret",
			AdminPermissions: bitmask.NewPermissionBitmask(123),
		}

		jsonBytes, err := json.Marshal(&exampleInput)
		require.NoError(t, err)

		var dest User
		assert.NoError(t, json.Unmarshal(jsonBytes, &dest))
	})
}

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
			ID:               12345,
			IsAdmin:          true,
			AdminPermissions: bitmask.NewPermissionBitmask(1),
		}

		expected := &SessionInfo{
			UserID:           exampleInput.ID,
			UserIsAdmin:      exampleInput.IsAdmin,
			AdminPermissions: exampleInput.AdminPermissions,
		}
		actual := exampleInput.ToSessionInfo()

		assert.Equal(t, expected, actual)
	})
}
