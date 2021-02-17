package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkUserCreationEquality(t *testing.T, expected *types.NewUserCreationInput, actual *types.UserCreationResponse) {
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
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be creatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			testClient := buildSimpleClient()

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			createdUser, err := testClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, createdUser, err)

			// Assert user equality.
			checkUserCreationEquality(t, exampleUserInput, createdUser)

			// Clean up.
			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			auditLogEntries, err := adminClient.GetAuditLogForUser(ctx, createdUser.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdUser.ID, audit.UserAssignmentKey)

			assert.NoError(t, adminClient.ArchiveUser(ctx, createdUser.ID))
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to read something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			// Fetch user.
			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			actual, err := adminClient.GetUser(ctx, nonexistentID)
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		subtest.Run("it should be readable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			createdUser, _, _ := createUserAndClientForTest(ctx, t)

			// Fetch user.
			adminClientLock.Lock()
			defer adminClientLock.Unlock()

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

	test.Run("Searching", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return empty slice when searching for a username that doesn'subtest exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			// Search For user.
			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			actual, err := adminClient.SearchForUsersByUsername(ctx, "   this is a really long string that contains characters unlikely to yield any real results   ")
			assert.Nil(t, actual)
			assert.NoError(t, err)
		})

		subtest.Run("it should only be accessible to admins", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			_, _, testClient := createUserAndClientForTest(ctx, t)

			// Search For user.
			actual, err := testClient.SearchForUsersByUsername(ctx, "   this is a really long string that contains characters unlikely to yield any real results   ")
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		subtest.Run("it should be searchable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
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
			adminClientLock.Lock()
			defer adminClientLock.Unlock()
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

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			testClient := buildSimpleClient()

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			createdUser, err := testClient.CreateUser(ctx, exampleUserInput)
			assert.NoError(t, err)
			assert.NotNil(t, createdUser)

			if createdUser == nil || err != nil {
				t.Log("something has gone awry, user returned is nil")
				t.FailNow()
			}

			// Execute.
			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			assert.NoError(t, adminClient.ArchiveUser(ctx, createdUser.ID))

			auditLogEntries, err := adminClient.GetAuditLogForUser(ctx, createdUser.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.UserArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdUser.ID, audit.UserAssignmentKey)
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			input := fakes.BuildFakeAccountStatusUpdateInput()
			input.NewReputation = types.BannedAccountStatus
			input.TargetAccountID = nonexistentID

			// Ban user.
			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			assert.Error(t, adminClient.UpdateAccountStatus(ctx, input))

			x, err := adminClient.GetAuditLogForUser(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		subtest.Run("it should be auditable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			testClient := buildSimpleClient()

			// Create user.
			exampleUser := fakes.BuildFakeUser()
			exampleUserInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
			createdUser, err := testClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, createdUser, err)

			// fetch audit log entries
			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			auditLogEntries, err := adminClient.GetAuditLogForUser(ctx, createdUser.ID)
			assert.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.AccountCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, 0, "")

			// Clean up item.
			assert.NoError(t, adminClient.ArchiveUser(ctx, createdUser.ID))
		})

		subtest.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			testClient := buildSimpleClient()

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
			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			assert.NoError(t, adminClient.ArchiveUser(ctx, createdUser.ID))
		})
	})
}
