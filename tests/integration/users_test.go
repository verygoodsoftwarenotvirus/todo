package integration

import (
	"fmt"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

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

func (s *TestSuite) TestUsersCreating() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be creatable via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			createdUser, err := testClients.main.CreateUser(ctx, exampleUserInput)
			requireNotNilAndNoProblems(t, createdUser, err)

			// Assert user equality.
			checkUserCreationEquality(t, exampleUserInput, createdUser)

			// Clean up.
			auditLogEntries, err := testClients.admin.GetAuditLogForUser(ctx, createdUser.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdUser.ID, audit.UserAssignmentKey)

			assert.NoError(t, testClients.admin.ArchiveUser(ctx, createdUser.ID))
		})
	}
}

func (s *TestSuite) TestUsersReading() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should return an error when trying to read a user that does not exist via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.admin.GetUser(ctx, nonexistentID)
			assert.Nil(t, actual)
			assert.Error(t, err)
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be read via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			user, _, _, _ := createUserAndClientForTest(ctx, t)

			actual, err := testClients.admin.GetUser(ctx, user.ID)
			if err != nil {
				t.Logf("error encountered trying to fetch user %q: %v\n", user.Username, err)
			}
			requireNotNilAndNoProblems(t, actual, err)

			// Assert user equality.
			checkUserEquality(t, user, actual)

			// Clean up.
			assert.NoError(t, testClients.admin.ArchiveUser(ctx, actual.ID))
		})
	}
}

func (s *TestSuite) TestUsersSearching() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("it should return empty slice when searching for a username that does not exist via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.admin.SearchForUsersByUsername(ctx, "   this is a really long string that contains characters unlikely to yield any real results   ")
			assert.Nil(t, actual)
			assert.NoError(t, err)
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("it should only be accessible to admins via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Search For user.
			actual, err := testClients.main.SearchForUsersByUsername(ctx, s.user.Username)
			assert.Nil(t, actual)
			assert.Error(t, err)
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("it should return be searchable via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleUsername := fakes.BuildFakeUser().Username

			// create users
			createdUserIDs := []uint64{}
			for i := 0; i < 5; i++ {
				user, err := utils.CreateServiceUser(ctx, urlToUse, fmt.Sprintf("%s%d", exampleUsername, i))
				require.NoError(t, err)
				createdUserIDs = append(createdUserIDs, user.ID)
			}

			// execute search
			actual, err := testClients.admin.SearchForUsersByUsername(ctx, exampleUsername)
			assert.NoError(t, err)
			assert.NotEmpty(t, actual)

			// ensure results look how we expect them to look
			for _, result := range actual {
				assert.True(t, strings.HasPrefix(result.Username, exampleUsername))
			}

			// clean up
			for _, id := range createdUserIDs {
				require.NoError(t, testClients.admin.ArchiveUser(ctx, id))
			}
		})
	}
}

func (s *TestSuite) TestUsersArchiving() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should fail to archive a non-existent user via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			assert.Error(t, testClients.admin.ArchiveUser(ctx, nonexistentID))
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be archived via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create user.
			exampleUserInput := fakes.BuildFakeUserCreationInput()
			createdUser, err := testClients.main.CreateUser(ctx, exampleUserInput)
			assert.NoError(t, err)
			assert.NotNil(t, createdUser)

			if createdUser == nil || err != nil {
				t.Log("something has gone awry, user returned is nil")
				t.FailNow()
			}

			// Execute.
			assert.NoError(t, testClients.admin.ArchiveUser(ctx, createdUser.ID))

			auditLogEntries, err := testClients.admin.GetAuditLogForUser(ctx, createdUser.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.UserArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdUser.ID, audit.UserAssignmentKey)
		})
	}
}

func (s *TestSuite) TestUsersAuditing() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("it should return an error when trying to audit something that does not exist via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			input := fakes.BuildFakeAccountStatusUpdateInput()
			input.NewReputation = types.BannedAccountStatus
			input.TargetAccountID = nonexistentID

			// Ban user.
			assert.Error(t, testClients.admin.UpdateAccountStatus(ctx, input))

			x, err := testClients.admin.GetAuditLogForUser(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("it should not be auditable by a non-admin via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create user.
			exampleUser := fakes.BuildFakeUser()
			exampleUserInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
			createdUser, err := testClients.main.CreateUser(ctx, exampleUserInput)
			requireNotNilAndNoProblems(t, createdUser, err)

			// fetch audit log entries
			actual, err := testClients.main.GetAuditLogForUser(ctx, createdUser.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClients.admin.ArchiveUser(ctx, createdUser.ID))
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be audited via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create user.
			exampleUser := fakes.BuildFakeUser()
			exampleUserInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
			createdUser, err := testClients.main.CreateUser(ctx, exampleUserInput)
			requireNotNilAndNoProblems(t, createdUser, err)

			// fetch audit log entries
			auditLogEntries, err := testClients.admin.GetAuditLogForUser(ctx, createdUser.ID)
			assert.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.AccountCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, 0, "")

			// Clean up item.
			assert.NoError(t, testClients.admin.ArchiveUser(ctx, createdUser.ID))
		})
	}
}

func (s *TestSuite) TestUsersAvatarManagement() {
	for a, c := range s.eachClient(pasetoAuthType) {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to upload an avatar via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			avatar := testutil.BuildArbitraryImagePNGBytes(256)

			require.NoError(t, testClients.main.UploadAvatarFromFile(ctx, avatar, "png"))

			// Assert user equality.
			user, err := testClients.admin.GetUser(ctx, s.user.ID)
			requireNotNilAndNoProblems(t, user, err)

			assert.NotEmpty(t, user.AvatarSrc)

			auditLogEntries, err := testClients.admin.GetAuditLogForUser(ctx, s.user.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.UserCreationEvent},
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.UserVerifyTwoFactorSecretEvent},
				{EventType: audit.SuccessfulLoginEvent},
				{EventType: audit.UserUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, s.user.ID, "")

			assert.NoError(t, testClients.admin.ArchiveUser(ctx, s.user.ID))
		})
	}
}
