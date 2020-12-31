package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAccountStatusUpdateInput builds a faked ItemCreationInput.
func BuildFakeAccountStatusUpdateInput() *types.AccountStatusUpdateInput {
	return &types.AccountStatusUpdateInput{
		TargetAccountID: uint64(fake.Uint32()),
		NewStatus:       types.GoodStandingAccountStatus,
		Reason:          fake.Sentence(10),
	}
}
