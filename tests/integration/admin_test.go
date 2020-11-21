package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin(test *testing.T) {
	test.SkipNow()

	test.Run("User Management", func(t *testing.T) {
		t.Run("users should be bannable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			actual, err := todoClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, actual, err)

			_, obligatoryCheckErr := todoClient.GetItems(ctx, nil)
			require.NoError(t, obligatoryCheckErr)

			// Assert user equality.
			checkUserCreationEquality(t, exampleUserInput, actual)

			assert.NoError(t, adminClient.BanUser(ctx, actual.ID))

			// Clean up.
			assert.NoError(t, todoClient.ArchiveUser(ctx, actual.ID))
		})
	})
}
