package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"
)

// BuildFakeAccountUserMembership builds a faked AccountUserMembership.
func BuildFakeAccountUserMembership() *types.AccountUserMembership {
	return &types.AccountUserMembership{
		ID:               ksuid.New().String(),
		BelongsToUser:    fake.UUID(),
		BelongsToAccount: fake.UUID(),
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

// BuildFakeAccountUserMembershipUpdateInputFromAccountUserMembership builds a faked AccountUserMembershipUpdateInput from an account user membership.
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

// BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership builds a faked AccountUserMembershipCreationInput from an account user membership.
func BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership(accountUserMembership *types.AccountUserMembership) *types.AccountUserMembershipCreationInput {
	return &types.AccountUserMembershipCreationInput{
		ID:               accountUserMembership.ID,
		BelongsToUser:    accountUserMembership.BelongsToUser,
		BelongsToAccount: accountUserMembership.BelongsToAccount,
	}
}
