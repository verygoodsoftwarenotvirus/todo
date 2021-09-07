package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserReputationUpdateInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &UserReputationUpdateInput{
			NewReputation: GoodStandingAccountStatus,
			Reason:        t.Name(),
			TargetUserID:  "123",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
