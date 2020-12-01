package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkUserCreationEquality(t *testing.T, expected *types.UserCreationInput, actual *types.UserCreationResponse) {
	t.Helper()

	twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(actual.TwoFactorQRCode)
	assert.NoError(t, err)

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.NotEmpty(t, twoFactorSecret)
	assert.NotZero(t, actual.CreatedOn)
}

func checkUserEquality(t *testing.T, expected, actual *types.User) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.LastUpdatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func TestUsers(test *testing.T) {
	test.Run("Creating", func(t *testing.T) {
		t.Run("should be creatable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testClient := buildSimpleClient(ctx, t)

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			actual, err := testClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, actual, err)

			// Assert user equality.
			checkUserCreationEquality(t, exampleUserInput, actual)

			// Clean up.
			assert.NoError(t, adminClient.ArchiveUser(ctx, actual.ID))
		})
	})

	test.Run("Reading", func(t *testing.T) {
		t.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			// Fetch user.
			actual, err := adminClient.GetUser(ctx, nonexistentID)
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		t.Run("it should be readable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			createdUser, _ := createUserAndClientForTest(ctx, t)

			// Fetch user.
			actual, err := adminClient.GetUser(ctx, createdUser.ID)
			if err != nil {
				t.Logf("error encountered trying to fetch user %q: %v\n", createdUser.Username, err)
			}
			checkValueAndError(t, actual, err)

			// Assert user equality.
			checkUserEquality(t, createdUser, actual)

			// Clean up.
			assert.NoError(t, adminClient.ArchiveUser(ctx, actual.ID))
		})
	})

	test.Run("Searching", func(t *testing.T) {
		t.Run("it should return empty slice when searching for a username that doesn't exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			// Search For user.
			actual, err := adminClient.SearchForUsersByUsername(ctx, "   this is a really long string that contains characters unlikely to yield any real results   ")
			assert.Nil(t, actual)
			assert.NoError(t, err)
		})

		t.Run("it should only be accessible to admins", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Search For user.
			actual, err := testClient.SearchForUsersByUsername(ctx, "   this is a really long string that contains characters unlikely to yield any real results   ")
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		t.Run("it should be searchable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			exampleUsername := fakes.BuildFakeUser().Username

			// create users
			createdUserIDs := []uint64{}
			for i := 0; i < 5; i++ {
				user, err := testutil.CreateServiceUser(ctx, urlToUse, fmt.Sprintf("%s%d", exampleUsername, i), debug)
				require.NoError(t, err)
				createdUserIDs = append(createdUserIDs, user.ID)
			}

			// execute search
			actual, err := adminClient.SearchForUsersByUsername(ctx, exampleUsername)
			assert.NoError(t, err)
			assert.NotEmpty(t, actual)

			// ensure results look how we expect them to look
			for _, result := range actual {
				assert.True(t, strings.HasPrefix(result.Username, exampleUsername))
			}

			// clean up
			for _, id := range createdUserIDs {
				require.NoError(t, adminClient.ArchiveUser(ctx, id))
			}
		})
	})

	test.Run("Deleting", func(t *testing.T) {
		t.Run("should be able to be deleted", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testClient := buildSimpleClient(ctx, t)

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			u, err := testClient.CreateUser(ctx, exampleUserInput)
			assert.NoError(t, err)
			assert.NotNil(t, u)

			if u == nil || err != nil {
				t.Log("something has gone awry, user returned is nil")
				t.FailNow()
			}

			// Execute.
			err = adminClient.ArchiveUser(ctx, u.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Auditing", func(t *testing.T) {
		t.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			input := fakes.BuildFakeAccountStatusUpdateInput()
			input.NewStatus = types.BannedAccountStatus
			input.TargetAccountID = nonexistentID

			// Ban user.
			assert.Error(t, adminClient.UpdateAccountStatus(ctx, input))

			x, err := adminClient.GetAuditLogForUser(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		t.Run("it should be auditable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testClient := buildSimpleClient(ctx, t)

			// Create user.
			exampleUser := fakes.BuildFakeUser()
			exampleUserInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
			createdUser, err := testClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, createdUser, err)

			// fetch audit log entries
			actual, err := adminClient.GetAuditLogForUser(ctx, createdUser.ID)
			assert.NoError(t, err)
			assert.Len(t, actual, 1)

			// Clean up item.
			assert.NoError(t, adminClient.ArchiveUser(ctx, createdUser.ID))
		})

		t.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testClient := buildSimpleClient(ctx, t)

			// Create user.
			exampleUser := fakes.BuildFakeUser()
			exampleUserInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
			createdUser, err := testClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, createdUser, err)

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForUser(ctx, createdUser.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, adminClient.ArchiveUser(ctx, createdUser.ID))
		})
	})
}
