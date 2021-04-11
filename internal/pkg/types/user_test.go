package types

import (
	"encoding/json"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_JSONUnmarshal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		exampleInput := User{
			Username:               "new_username",
			HashedPassword:         "updated_hashed_pass",
			TwoFactorSecret:        "new fancy secret",
			ServiceAdminPermission: permissions.NewServiceAdminPermissions(123),
		}

		jsonBytes, err := json.Marshal(&exampleInput)
		require.NoError(t, err)

		var dest User
		assert.NoError(t, json.Unmarshal(jsonBytes, &dest))
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

func TestRequestContextFromUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := &User{
			ID:                     12345,
			ServiceAdminPermission: permissions.NewServiceAdminPermissions(1),
		}

		exampleAccount := &Account{
			ID:            54321,
			BelongsToUser: exampleUser.ID,
		}

		examplePermissions := map[uint64]*UserAccountMembershipInfo{}

		expected := &RequestContext{
			Requester: RequesterInfo{
				ID:                     exampleUser.ID,
				ServiceAdminPermission: exampleUser.ServiceAdminPermission,
			},
			ActiveAccountID:       exampleAccount.ID,
			AccountPermissionsMap: examplePermissions,
		}

		actual, _ := RequestContextFromUser(exampleUser, exampleAccount.ID, examplePermissions)

		assert.Equal(t, expected, actual)
	})
}
