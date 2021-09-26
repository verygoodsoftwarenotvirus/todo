package fakes

import (
	"time"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// BuildFakeSessionContextData builds a faked SessionContextData.
func BuildFakeSessionContextData() *types.SessionContextData {
	fakeAccountID := fake.UUID()

	return &types.SessionContextData{
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			fakeAccountID: authorization.NewAccountRolePermissionChecker(authorization.AccountAdminRole.String()),
		},
		Requester: types.RequesterInfo{
			Reputation:            types.GoodStandingAccountStatus,
			ReputationExplanation: "",
			UserID:                ksuid.New().String(),
			ServicePermissions:    authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
		},
		ActiveAccountID: fakeAccountID,
	}
}

// BuildFakeSessionContextDataForAccount builds a faked SessionContextData.
func BuildFakeSessionContextDataForAccount(account *types.Account) *types.SessionContextData {
	fakeAccountID := fake.UUID()

	return &types.SessionContextData{
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			account.ID: authorization.NewAccountRolePermissionChecker(authorization.ServiceUserRole.String()),
		},
		Requester: types.RequesterInfo{
			Reputation:            types.GoodStandingAccountStatus,
			ReputationExplanation: "",
			UserID:                ksuid.New().String(),
			ServicePermissions:    authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
		},
		ActiveAccountID: fakeAccountID,
	}
}

// BuildFakeAddUserToAccountInput builds a faked AddUserToAccountInput.
func BuildFakeAddUserToAccountInput() *types.AddUserToAccountInput {
	return &types.AddUserToAccountInput{
		Reason:       fake.Sentence(10),
		UserID:       ksuid.New().String(),
		AccountID:    ksuid.New().String(),
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
		CurrentOwner: fake.UUID(),
		NewOwner:     fake.UUID(),
	}
}

// BuildFakeChangeActiveAccountInput builds a faked ChangeActiveAccountInput.
func BuildFakeChangeActiveAccountInput() *types.ChangeActiveAccountInput {
	return &types.ChangeActiveAccountInput{
		AccountID: fake.UUID(),
	}
}

// BuildFakePASETOCreationInput builds a faked PASETOCreationInput.
func BuildFakePASETOCreationInput() *types.PASETOCreationInput {
	return &types.PASETOCreationInput{
		ClientID:    ksuid.New().String(),
		RequestTime: time.Now().Unix(),
	}
}
