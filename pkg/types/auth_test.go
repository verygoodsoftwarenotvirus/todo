package types

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"github.com/stretchr/testify/assert"
)

func TestChangeActiveAccountInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &ChangeActiveAccountInput{
			AccountID: 123,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestPASETOCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &PASETOCreationInput{
			ClientID:    t.Name(),
			RequestTime: time.Now().Unix(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestSessionContextData_ToBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &SessionContextData{}

		assert.NotEmpty(t, x.ToBytes())
	})
}

func TestSessionContextDataFromUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := &User{
			ID: 12345,
		}

		exampleAccount := &Account{
			ID: 54321,
		}

		examplePermissions := map[uint64]*UserAccountMembershipInfo{}
		exampleAccountPermissions := map[uint64]authorization.AccountRolePermissionsChecker{}

		expected := &SessionContextData{
			Requester: RequesterInfo{
				ID:                     exampleUser.ID,
				ServicePermissions:     authorization.NewServiceRolePermissionChecker(exampleUser.ServiceRoles...),
				RequiresPasswordChange: exampleUser.RequiresPasswordChange,
			},
			ActiveAccountID:       exampleAccount.ID,
			AccountPermissionsMap: examplePermissions,
			AccountRolesMap:       exampleAccountPermissions,
		}

		actual, err := SessionContextDataFromUser(exampleUser, exampleAccount.ID, examplePermissions, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("with nil user", func(t *testing.T) {
		t.Parallel()

		exampleAccount := &Account{ID: 54321}
		examplePermissions := map[uint64]*UserAccountMembershipInfo{}

		actual, err := SessionContextDataFromUser(nil, exampleAccount.ID, examplePermissions, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := &User{
			ID: 12345,
		}

		examplePermissions := map[uint64]*UserAccountMembershipInfo{}

		actual, err := SessionContextDataFromUser(exampleUser, 0, examplePermissions, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil permissions map", func(t *testing.T) {
		t.Parallel()

		exampleUser := &User{
			ID: 12345,
		}

		exampleAccount := &Account{
			ID: 54321,
		}

		_, err := SessionContextDataFromUser(exampleUser, exampleAccount.ID, nil, nil)
		assert.Error(t, err)
	})
}
