package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAccountUserMembership builds a faked item.
func BuildFakeAccountUserMembership() *types.AccountUserMembership {
	return &types.AccountUserMembership{
		ID:               uint64(fake.Uint32()),
		BelongsToUser:    fake.Uint64(),
		BelongsToAccount: fake.Uint64(),
		AccountRoles:     []string{authorization.AccountMemberRole.String()},
		CreatedOn:        0,
		ArchivedOn:       nil,
	}
}

// BuildFakeAccountUserMembershipList builds a faked AccountUserMembershipList.
func BuildFakeAccountUserMembershipList() *types.AccountUserMembershipList {
	var examples []*types.AccountUserMembership
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeAccountUserMembership())
	}

	return &types.AccountUserMembershipList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		AccountUserMemberships: examples,
	}
}

// BuildFakeAccountUserMembershipUpdateInputFromAccountUserMembership builds a faked AccountUserMembershipUpdateInput from an item.
func BuildFakeAccountUserMembershipUpdateInputFromAccountUserMembership(accountUserMembership *types.AccountUserMembership) *types.AccountUserMembershipUpdateInput {
	return &types.AccountUserMembershipUpdateInput{
		BelongsToUser:    accountUserMembership.BelongsToUser,
		BelongsToAccount: accountUserMembership.BelongsToAccount,
	}
}

// BuildFakeAccountUserMembershipCreationInput builds a faked AccountUserMembershipCreationInput.
func BuildFakeAccountUserMembershipCreationInput() *types.AccountUserMembershipCreationInput {
	return BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership(BuildFakeAccountUserMembership())
}

// BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership builds a faked AccountUserMembershipCreationInput from an item.
func BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership(accountUserMembership *types.AccountUserMembership) *types.AccountUserMembershipCreationInput {
	return &types.AccountUserMembershipCreationInput{
		BelongsToUser:    accountUserMembership.BelongsToUser,
		BelongsToAccount: accountUserMembership.BelongsToAccount,
	}
}
