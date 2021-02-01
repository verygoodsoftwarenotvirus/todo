package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAccount builds a faked account.
func BuildFakeAccount() *types.Account {
	return &types.Account{
		ID:            uint64(fake.Uint32()),
		ExternalID:    fake.UUID(),
		Name:          fake.Word(),
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		BelongsToUser: fake.Uint64(),
	}
}

// BuildFakeAccountList builds a faked AccountList.
func BuildFakeAccountList() *types.AccountList {
	var examples []*types.Account
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeAccount())
	}

	return &types.AccountList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Accounts: examples,
	}
}

// BuildFakeAccountUpdateInputFromAccount builds a faked AccountUpdateInput from an account.
func BuildFakeAccountUpdateInputFromAccount(account *types.Account) *types.AccountUpdateInput {
	return &types.AccountUpdateInput{
		Name:          account.Name,
		BelongsToUser: account.BelongsToUser,
	}
}

// BuildFakeAccountCreationInput builds a faked AccountCreationInput.
func BuildFakeAccountCreationInput() *types.AccountCreationInput {
	account := BuildFakeAccount()
	return BuildFakeAccountCreationInputFromAccount(account)
}

// BuildFakeAccountCreationInputFromAccount builds a faked AccountCreationInput from an account.
func BuildFakeAccountCreationInputFromAccount(account *types.Account) *types.AccountCreationInput {
	return &types.AccountCreationInput{
		Name:          account.Name,
		BelongsToUser: account.BelongsToUser,
	}
}
