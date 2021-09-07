package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"
)

// BuildFakeAccount builds a faked account.
func BuildFakeAccount() *types.Account {
	return &types.Account{
		ID:                         ksuid.New().String(),
		Name:                       fake.UUID(),
		BillingStatus:              types.PaidAccountBillingStatus,
		ContactEmail:               fake.Email(),
		ContactPhone:               fake.PhoneFormatted(),
		PaymentProcessorCustomerID: fake.UUID(),
		CreatedOn:                  uint64(uint32(fake.Date().Unix())),
		BelongsToUser:              fake.UUID(),
		Members:                    BuildFakeAccountUserMembershipList().AccountUserMemberships,
	}
}

// BuildFakeAccountForUser builds a faked account.
func BuildFakeAccountForUser(u *types.User) *types.Account {
	return &types.Account{
		ID:            ksuid.New().String(),
		Name:          u.Username,
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		BelongsToUser: u.ID,
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

// BuildFakeAccountUpdateInput builds a faked AccountUpdateInput from an account.
func BuildFakeAccountUpdateInput() *types.AccountUpdateInput {
	account := BuildFakeAccount()
	return &types.AccountUpdateInput{
		Name:          account.Name,
		BelongsToUser: account.BelongsToUser,
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
		ID:            ksuid.New().String(),
		Name:          account.Name,
		ContactEmail:  account.ContactEmail,
		ContactPhone:  account.ContactPhone,
		BelongsToUser: account.BelongsToUser,
	}
}
