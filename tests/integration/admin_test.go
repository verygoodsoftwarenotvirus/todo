package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin(test *testing.T) {
	test.Run("User Management", func(t *testing.T) {
		t.Run("users should be bannable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			user, testClient := createUserAndClientForTest(ctx, t)

			// Assert that user can access service
			_, initialCheckErr := testClient.GetItems(ctx, nil)
			require.NoError(t, initialCheckErr)

			assert.NoError(t, adminClient.BanUser(ctx, user.ID))

			// Assert user can no longer access service
			_, subsequentCheckErr := testClient.GetItems(ctx, nil)
			assert.Error(t, subsequentCheckErr)

			// Clean up.
			assert.NoError(t, todoClient.ArchiveUser(ctx, user.ID))
		})
	})
}
