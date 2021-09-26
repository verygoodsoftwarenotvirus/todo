package fakes

import (
	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// BuildFakeUserReputationUpdateInput builds a faked UserReputationUpdateInput.
func BuildFakeUserReputationUpdateInput() *types.UserReputationUpdateInput {
	return &types.UserReputationUpdateInput{
		TargetUserID:  ksuid.New().String(),
		NewReputation: types.GoodStandingAccountStatus,
		Reason:        fake.Sentence(10),
	}
}
