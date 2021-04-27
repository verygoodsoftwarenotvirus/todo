package types

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccountSubscriptionPlan_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &AccountSubscriptionPlan{}
		input := &AccountSubscriptionPlanUpdateInput{
			Name: t.Name(),
		}

		assert.NotEmpty(t, x.Update(input))
	})
}

func TestAccountSubscriptionPlanCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &AccountSubscriptionPlanCreationInput{
			Name:        t.Name(),
			Description: t.Name(),
			Price:       1234,
			Period:      time.Hour,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
