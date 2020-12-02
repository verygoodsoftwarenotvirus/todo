package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin(test *testing.T) {
	test.Parallel()

	test.Run("User Management", func(t *testing.T) {
		t.Parallel()

		t.Run("it should return an error when trying to ban a user that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			input := fakes.BuildFakeAccountStatusUpdateInput()
			input.TargetAccountID = nonexistentID

			// Ban user.
			assert.Error(t, adminClient.UpdateAccountStatus(ctx, input))
		})

		t.Run("users should be bannable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			user, testClient := createUserAndClientForTest(ctx, t)

			// Assert that user can access service
			_, initialCheckErr := testClient.GetItems(ctx, nil)
			require.NoError(t, initialCheckErr)

			input := &types.AccountStatusUpdateInput{
				TargetAccountID: user.ID,
				NewStatus:       types.BannedAccountStatus,
				Reason:          "testing",
			}

			assert.NoError(t, adminClient.UpdateAccountStatus(ctx, input))

			// Assert user can no longer access service
			_, subsequentCheckErr := testClient.GetItems(ctx, nil)
			assert.Error(t, subsequentCheckErr)

			// Clean up.
			assert.NoError(t, adminClient.ArchiveUser(ctx, user.ID))
		})
	})
}
