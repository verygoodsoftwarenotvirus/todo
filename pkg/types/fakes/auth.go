package fakes

import (
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeSessionContextData builds a faked SessionContextData.
func BuildFakeSessionContextData() *types.SessionContextData {
	fakeAccountID := fake.Uint64()

	return &types.SessionContextData{
		AccountPermissionsMap: map[uint64]*types.UserAccountMembershipInfo{
			fakeAccountID: {
				AccountName: fake.Name(),
				AccountID:   fakeAccountID,
				Permissions: permissions.NewServiceUserPermissions(fake.Int64()),
			},
		},
		Requester: types.RequesterInfo{
			Reputation:             types.GoodStandingAccountStatus,
			ReputationExplanation:  "",
			ID:                     fake.Uint64(),
			ServiceAdminPermission: permissions.NewServiceAdminPermissions(fake.Int64()),
			RequiresPasswordChange: false,
			ServiceRole:            authorization.ServiceUserRole,
		},
		ActiveAccountID: fakeAccountID,
	}
}

// BuildFakeSessionContextDataForAccount builds a faked SessionContextData.
func BuildFakeSessionContextDataForAccount(account *types.Account) *types.SessionContextData {
	fakeAccountID := fake.Uint64()

	return &types.SessionContextData{
		AccountPermissionsMap: map[uint64]*types.UserAccountMembershipInfo{
			account.ID: {
				AccountName: account.Name,
				AccountID:   account.ID,
				Permissions: permissions.NewServiceUserPermissions(fake.Int64()),
			},
		},
		Requester: types.RequesterInfo{
			Reputation:             types.GoodStandingAccountStatus,
			ReputationExplanation:  "",
			ID:                     fake.Uint64(),
			ServiceAdminPermission: permissions.NewServiceAdminPermissions(fake.Int64()),
			RequiresPasswordChange: false,
			ServiceRole:            authorization.ServiceUserRole,
		},
		ActiveAccountID: fakeAccountID,
	}
}

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
