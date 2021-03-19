package integration

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	testlogging "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/testing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
)

func checkAccountEquality(t *testing.T, expected, actual *types.Account) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected BucketName for account #%d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.NotZero(t, actual.CreatedOn)
}

func (s *TestSuite) TestAccountsCreating() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to create accounts via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create account.
			exampleAccount := fakes.BuildFakeAccount()
			exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
			createdAccount, err := testClients.main.CreateAccount(ctx, exampleAccountInput)
			requireNotNilAndNoProblems(t, createdAccount, err)

			// Assert account equality.
			checkAccountEquality(t, exampleAccount, createdAccount)

			auditLogEntries, err := testClients.admin.GetAuditLogForAccount(ctx, createdAccount.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccount.ID, audit.AccountAssignmentKey)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))
		})
	}
}

func (s *TestSuite) TestAccountsListing() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to list accounts via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create accounts.
			var expected []*types.Account
			for i := 0; i < 5; i++ {
				// Create account.
				exampleAccount := fakes.BuildFakeAccount()
				exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
				createdAccount, accountCreationErr := testClients.main.CreateAccount(ctx, exampleAccountInput)
				requireNotNilAndNoProblems(t, createdAccount, accountCreationErr)

				expected = append(expected, createdAccount)
			}

			// Assert account list equality.
			actual, err := testClients.main.GetAccounts(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Accounts),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Accounts),
			)

			// Clean up.
			for _, createdAccount := range actual.Accounts {
				assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))
			}
		})
	}
}

func (s *TestSuite) TestAccountsReading() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to read a non-existent account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Attempt to fetch nonexistent account.
			_, err := testClients.main.GetAccount(ctx, nonexistentID)
			assert.Error(t, err)
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to read an account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create account.
			exampleAccount := fakes.BuildFakeAccount()
			exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
			createdAccount, err := testClients.main.CreateAccount(ctx, exampleAccountInput)
			requireNotNilAndNoProblems(t, createdAccount, err)

			// Fetch account.
			actual, err := testClients.main.GetAccount(ctx, createdAccount.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert account equality.
			checkAccountEquality(t, exampleAccount, actual)

			// Clean up account.
			assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))
		})
	}
}

func (s *TestSuite) TestAccountsUpdating() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to update a non-existent account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleAccount := fakes.BuildFakeAccount()
			exampleAccount.ID = nonexistentID

			assert.Error(t, testClients.main.UpdateAccount(ctx, exampleAccount))
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to update an account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create account.
			exampleAccount := fakes.BuildFakeAccount()
			exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
			createdAccount, err := testClients.main.CreateAccount(ctx, exampleAccountInput)
			requireNotNilAndNoProblems(t, createdAccount, err)

			// Change account.
			createdAccount.Update(converters.ConvertAccountToAccountUpdateInput(exampleAccount))
			assert.NoError(t, testClients.main.UpdateAccount(ctx, createdAccount))

			// Fetch account.
			actual, err := testClients.main.GetAccount(ctx, createdAccount.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert account equality.
			checkAccountEquality(t, exampleAccount, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			auditLogEntries, err := testClients.admin.GetAuditLogForAccount(ctx, createdAccount.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.AccountUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccount.ID, audit.AccountAssignmentKey)

			// Clean up account.
			assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))
		})
	}
}

func (s *TestSuite) TestAccountsArchiving() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to archiv a non-existent account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			assert.Error(t, testClients.main.ArchiveAccount(ctx, nonexistentID))
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to archive an account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create account.
			exampleAccount := fakes.BuildFakeAccount()
			exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
			createdAccount, err := testClients.main.CreateAccount(ctx, exampleAccountInput)
			requireNotNilAndNoProblems(t, createdAccount, err)

			// Clean up account.
			assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))

			auditLogEntries, err := testClients.admin.GetAuditLogForAccount(ctx, createdAccount.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.AccountArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccount.ID, audit.AccountAssignmentKey)
		})
	}
}

func (s *TestSuite) TestAccountsMemberships() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to change members of an account via %s", authType), func() {
			t := s.T()

			logger := testlogging.NewLogger(t)

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			const userCount = 1
			// create dummy users
			users := []*types.User{}
			clients := []*httpclient.Client{}

			// create users
			for i := 0; i < userCount; i++ {
				u, _, c, _ := createUserAndClientForTest(ctx, t)
				users = append(users, u)
				clients = append(clients, c)
			}

			// fetch account data
			accountCreationInput := &types.AccountCreationInput{
				Name:                   fakes.BuildFakeAccount().Name,
				DefaultUserPermissions: permissions.ServiceUserPermissions(math.MaxUint32),
			}
			account, accountCreationErr := testClients.main.CreateAccount(ctx, accountCreationInput)
			require.NoError(t, accountCreationErr)
			require.NotNil(t, account)

			require.Equal(t, accountCreationInput.DefaultUserPermissions, account.DefaultUserPermissions, "expected and actual permissions do not match")
			require.NoError(t, testClients.main.SwitchActiveAccount(ctx, account.ID))

			logger = logger.WithValue(keys.AccountIDKey, account.ID)
			logger.Debug("switched account")

			// create a webhook
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhook, creationErr := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, createdWebhook, creationErr)
			require.Equal(t, account.ID, createdWebhook.BelongsToAccount)

			logger = logger.WithValue(keys.WebhookIDKey, createdWebhook.ID)
			logger.Debug(fmt.Sprintf("created webhook for account %d", createdWebhook.BelongsToAccount))

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.WebhookCreationEvent},
			}

			// check that each user cannot see the webhook
			for i := 0; i < userCount; i++ {
				webhook, err := clients[i].GetWebhook(ctx, createdWebhook.ID)
				require.Nil(t, webhook)
				require.Error(t, err)
			}

			// add them to the account
			for i := 0; i < userCount; i++ {
				require.NoError(t, testClients.main.AddUserToAccount(ctx, account.ID, &types.AddUserToAccountInput{
					UserID: users[i].ID,
					Reason: t.Name(),
				}))
				expectedAuditLogEntries = append(expectedAuditLogEntries, &types.AuditLogEntry{EventType: audit.UserAddedToAccountEvent})
				require.NoError(t, clients[i].SwitchActiveAccount(ctx, account.ID))
			}

			// check that each user can see the webhook
			for i := 0; i < userCount; i++ {
				webhook, err := clients[i].GetWebhook(ctx, createdWebhook.ID)
				requireNotNilAndNoProblems(t, webhook, err)
			}

			// check that each user cannot update the webhook
			for i := 0; i < userCount; i++ {
				require.Error(t, clients[i].UpdateWebhook(ctx, createdWebhook))
			}

			// grant all permissions
			for i := 0; i < userCount; i++ {
				input := &types.ModifyUserPermissionsInput{
					UserAccountPermissions: testutil.BuildMaxUserPerms(),
					Reason:                 t.Name(),
				}
				require.NoError(t, testClients.main.ModifyMemberPermissions(ctx, account.ID, users[i].ID, input))
				expectedAuditLogEntries = append(expectedAuditLogEntries, &types.AuditLogEntry{EventType: audit.UserAccountPermissionsModifiedEvent})
			}

			// check that each user can update the webhook
			for i := 0; i < userCount; i++ {
				createdWebhook.Name = fmt.Sprintf("%s_%d", createdWebhook.Name, time.Now().UnixNano())
				require.NoError(t, clients[i].UpdateWebhook(ctx, createdWebhook))
				expectedAuditLogEntries = append(expectedAuditLogEntries, &types.AuditLogEntry{EventType: audit.WebhookUpdateEvent})
			}

			// remove users from account
			for i := 0; i < userCount; i++ {
				require.NoError(t, testClients.main.RemoveUserFromAccount(ctx, account.ID, users[i].ID, t.Name()))
			}

			// check that each user cannot see the webhook
			for i := 0; i < userCount; i++ {
				webhook, err := clients[i].GetWebhook(ctx, createdWebhook.ID)
				require.Nil(t, webhook)
				require.Error(t, err)
			}

			// check audit log entries
			auditLogEntries, err := testClients.admin.GetAuditLogForAccount(ctx, account.ID)
			require.NoError(t, err)

			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, account.ID, audit.AccountAssignmentKey)

			// Clean up.
			require.NoError(t, testClients.main.ArchiveWebhook(ctx, createdWebhook.ID))

			for i := 0; i < userCount; i++ {
				require.NoError(t, testClients.admin.ArchiveUser(ctx, users[i].ID))
			}
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to transfer ownership of an account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create users
			futureOwner, _, _, futureOwnerClient := createUserAndClientForTest(ctx, t)

			// fetch account data
			accountCreationInput := &types.AccountCreationInput{
				Name:                   fakes.BuildFakeAccount().Name,
				DefaultUserPermissions: permissions.ServiceUserPermissions(math.MaxUint32),
			}
			account, accountCreationErr := testClients.main.CreateAccount(ctx, accountCreationInput)
			require.NoError(t, accountCreationErr)
			require.NotNil(t, account)

			t.Logf("created account #%d", account.ID)

			require.Equal(t, accountCreationInput.DefaultUserPermissions, account.DefaultUserPermissions, "expected and actual permissions do not match")
			require.NoError(t, testClients.main.SwitchActiveAccount(ctx, account.ID))

			t.Logf("switched to active account: %d", account.ID)

			// create a webhook
			exampleWebhook := fakes.BuildFakeWebhook()
			exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
			createdWebhook, creationErr := testClients.main.CreateWebhook(ctx, exampleWebhookInput)
			requireNotNilAndNoProblems(t, createdWebhook, creationErr)

			t.Logf("created webhook #%d belonging to account #%d", createdWebhook.ID, createdWebhook.BelongsToAccount)
			require.Equal(t, account.ID, createdWebhook.BelongsToAccount)

			// check that user cannot see the webhook
			webhook, err := futureOwnerClient.GetWebhook(ctx, createdWebhook.ID)
			require.Nil(t, webhook)
			require.Error(t, err)

			// add them to the account
			require.NoError(t, testClients.main.TransferAccountOwnership(ctx, account.ID, &types.TransferAccountOwnershipInput{
				Reason:       t.Name(),
				CurrentOwner: account.BelongsToUser,
				NewOwner:     futureOwner.ID,
			}))

			t.Logf("transferred account %d from user %d to user %d", account.ID, account.BelongsToUser, futureOwner.ID)

			require.NoError(t, futureOwnerClient.SwitchActiveAccount(ctx, account.ID))

			// check that user can see the webhook
			webhook, err = futureOwnerClient.GetWebhook(ctx, createdWebhook.ID)
			requireNotNilAndNoProblems(t, webhook, err)

			// check that old user cannot see the webhook
			webhook, err = testClients.main.GetWebhook(ctx, createdWebhook.ID)
			require.Nil(t, webhook)
			require.Error(t, err)

			// check that new owner can update the webhook
			require.NoError(t, futureOwnerClient.UpdateWebhook(ctx, createdWebhook))

			// check audit log entries
			auditLogEntries, err := testClients.admin.GetAuditLogForAccount(ctx, account.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.AccountCreationEvent},
				{EventType: audit.UserAddedToAccountEvent},
				{EventType: audit.WebhookCreationEvent},
				{EventType: audit.AccountTransferredEvent},
				{EventType: audit.WebhookUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, account.ID, audit.AccountAssignmentKey)

			// Clean up.
			require.Error(t, testClients.main.ArchiveWebhook(ctx, createdWebhook.ID))
			require.NoError(t, futureOwnerClient.ArchiveWebhook(ctx, createdWebhook.ID))

			require.NoError(t, testClients.admin.ArchiveUser(ctx, futureOwner.ID))
		})
	}
}

func (s *TestSuite) TestAccountsAuditing() {
	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to audit a non-existent account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			x, err := testClients.admin.GetAuditLogForAccount(ctx, nonexistentID)

			assert.NoError(t, err)
			assert.Empty(t, x)
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to audit an account as non-admin via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create account.
			exampleAccount := fakes.BuildFakeAccount()
			exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
			createdAccount, err := testClients.main.CreateAccount(ctx, exampleAccountInput)
			requireNotNilAndNoProblems(t, createdAccount, err)

			// fetch audit log entries
			actual, err := testClients.main.GetAuditLogForAccount(ctx, createdAccount.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up account.
			assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))
		})
	}

	for a, c := range s.eachClientExcept() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to audit an account via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create account.
			exampleAccount := fakes.BuildFakeAccount()
			exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
			createdAccount, err := testClients.main.CreateAccount(ctx, exampleAccountInput)
			requireNotNilAndNoProblems(t, createdAccount, err)

			// fetch audit log entries
			actual, err := testClients.admin.GetAuditLogForAccount(ctx, createdAccount.ID)
			assert.NoError(t, err)
			assert.NotNil(t, actual)

			// Clean up account.
			assert.NoError(t, testClients.main.ArchiveAccount(ctx, createdAccount.ID))
		})
	}
}
