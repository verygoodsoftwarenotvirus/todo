package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeUserReputationUpdateInput builds a faked ItemCreationInput.
func BuildFakeUserReputationUpdateInput() *types.UserReputationUpdateInput {
	return &types.UserReputationUpdateInput{
		TargetUserID:  uint64(fake.Uint32()),
		NewReputation: types.GoodStandingAccountStatus,
		Reason:        fake.Sentence(10),
	}
}
