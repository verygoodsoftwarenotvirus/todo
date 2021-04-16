package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAddUserToAccountInput builds a faked AddUserToAccountInput.
func BuildFakeAddUserToAccountInput() *types.AddUserToAccountInput {
	return &types.AddUserToAccountInput{
		Reason:                 fake.Sentence(10),
		UserID:                 fake.Uint64(),
		AccountID:              fake.Uint64(),
		UserAccountPermissions: permissions.ServiceUserPermission(fake.Int64()),
	}
}

// BuildFakeUserPermissionModificationInput builds a faked ModifyUserPermissionsInput.
func BuildFakeUserPermissionModificationInput() *types.ModifyUserPermissionsInput {
	return &types.ModifyUserPermissionsInput{
		Reason:                 fake.Sentence(10),
		UserAccountPermissions: permissions.ServiceUserPermission(fake.Int64()),
	}
}

// BuildFakeTransferAccountOwnershipInput builds a faked TransferAccountOwnershipInput.
func BuildFakeTransferAccountOwnershipInput() *types.TransferAccountOwnershipInput {
	return &types.TransferAccountOwnershipInput{
		Reason:       fake.Sentence(10),
		CurrentOwner: fake.Uint64(),
		NewOwner:     fake.Uint64(),
	}
}

// BuildFakeChangeActiveAccountInput builds a faked BuildFakeChangeActiveAccountInput.
func BuildFakeChangeActiveAccountInput() *types.ChangeActiveAccountInput {
	return &types.ChangeActiveAccountInput{
		AccountID: fake.Uint64(),
	}
}
