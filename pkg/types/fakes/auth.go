package fakes

import (
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeSessionContextData builds a faked SessionContextData.
func BuildFakeSessionContextData() *types.SessionContextData {
	fakeAccountID := fake.Uint64()

	return &types.SessionContextData{
		AccountPermissions: map[uint64]authorization.AccountRolePermissionsChecker{
			fakeAccountID: authorization.NewAccountRolePermissionChecker(authorization.AccountAdminRole.String()),
		},
		Requester: types.RequesterInfo{
			Reputation:            types.GoodStandingAccountStatus,
			ReputationExplanation: "",
			UserID:                fake.Uint64(),
			ServicePermissions:    authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
		},
		ActiveAccountID: fakeAccountID,
	}
}

// BuildFakeSessionContextDataForAccount builds a faked SessionContextData.
func BuildFakeSessionContextDataForAccount(account *types.Account) *types.SessionContextData {
	fakeAccountID := fake.Uint64()

	return &types.SessionContextData{
		AccountPermissions: map[uint64]authorization.AccountRolePermissionsChecker{
			account.ID: authorization.NewAccountRolePermissionChecker(authorization.ServiceUserRole.String()),
		},
		Requester: types.RequesterInfo{
			Reputation:            types.GoodStandingAccountStatus,
			ReputationExplanation: "",
			UserID:                fake.Uint64(),
			ServicePermissions:    authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
		},
		ActiveAccountID: fakeAccountID,
	}
}

// BuildFakeAddUserToAccountInput builds a faked AddUserToAccountInput.
func BuildFakeAddUserToAccountInput() *types.AddUserToAccountInput {
	return &types.AddUserToAccountInput{
		Reason:       fake.Sentence(10),
		UserID:       fake.Uint64(),
		AccountID:    fake.Uint64(),
		AccountRoles: []string{authorization.AccountMemberRole.String()},
	}
}

// BuildFakeUserPermissionModificationInput builds a faked ModifyUserPermissionsInput.
func BuildFakeUserPermissionModificationInput() *types.ModifyUserPermissionsInput {
	return &types.ModifyUserPermissionsInput{
		Reason:   fake.Sentence(10),
		NewRoles: []string{authorization.AccountMemberRole.String()},
	}
}

// BuildFakeTransferAccountOwnershipInput builds a faked AccountOwnershipTransferInput.
func BuildFakeTransferAccountOwnershipInput() *types.AccountOwnershipTransferInput {
	return &types.AccountOwnershipTransferInput{
		Reason:       fake.Sentence(10),
		CurrentOwner: fake.Uint64(),
		NewOwner:     fake.Uint64(),
	}
}

// BuildFakeChangeActiveAccountInput builds a faked ChangeActiveAccountInput.
func BuildFakeChangeActiveAccountInput() *types.ChangeActiveAccountInput {
	return &types.ChangeActiveAccountInput{
		AccountID: fake.Uint64(),
	}
}

// BuildFakePASETOCreationInput builds a faked PASETOCreationInput.
func BuildFakePASETOCreationInput() *types.PASETOCreationInput {
	return &types.PASETOCreationInput{
		ClientID:    fake.UUID(),
		RequestTime: time.Now().Unix(),
	}
}
