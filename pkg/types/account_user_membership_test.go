package types

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

	"github.com/stretchr/testify/assert"
)

func TestAddUserToAccountInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &AddUserToAccountInput{
			UserID: 123,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestTransferAccountOwnershipInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &TransferAccountOwnershipInput{
			CurrentOwner: 123,
			NewOwner:     321,
			Reason:       t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestModifyUserPermissionsInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &ModifyUserPermissionsInput{
			UserAccountPermissions: permissions.ServiceUserPermission(123),
			Reason:                 t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
