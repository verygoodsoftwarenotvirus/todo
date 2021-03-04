package integration

import (
	"context"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"github.com/stretchr/testify/assert"
)

func checkAccountEquality(t *testing.T, expected, actual *types.Account) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected BucketName for account #%d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.NotZero(t, actual.CreatedOn)
}

func TestAccounts(test *testing.T) {
	test.Parallel()

	/*
		test.Run("Creating", func(subtest *testing.T) {
			subtest.Parallel()

			runTestForAllAuthMethods(subtest, "should be creatable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create account.
					exampleAccount := fakes.BuildFakeAccount()
					exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
					createdAccount, err := testClient.CreateAccount(ctx, exampleAccountInput)
					checkValueAndError(t, createdAccount, err)

					// Assert account equality.
					checkAccountEquality(t, exampleAccount, createdAccount)

					adminClientLock.Lock()
					defer adminClientLock.Unlock()

					auditLogEntries, err := adminCookieClient.GetAuditLogForAccount(ctx, createdAccount.ID)
					require.NoError(t, err)

					expectedAuditLogEntries := []*types.AuditLogEntry{
						{EventType: audit.AccountCreationEvent},
					}
					validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccount.ID, audit.AccountAssignmentKey)

					// Clean up.
					assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
				}
			})
		})

		test.Run("Listing", func(subtest *testing.T) {
			subtest.Parallel()

			runTestForAllAuthMethods(subtest, "should be able to be read in a list", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create accounts.
					var expected []*types.Account
					for i := 0; i < 5; i++ {
						// Create account.
						exampleAccount := fakes.BuildFakeAccount()
						exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
						createdAccount, accountCreationErr := testClient.CreateAccount(ctx, exampleAccountInput)
						checkValueAndError(t, createdAccount, accountCreationErr)

						expected = append(expected, createdAccount)
					}

					// Assert account list equality.
					actual, err := testClient.GetAccounts(ctx, nil)
					checkValueAndError(t, actual, err)
					assert.True(
						t,
						len(expected) <= len(actual.Accounts),
						"expected %d to be <= %d",
						len(expected),
						len(actual.Accounts),
					)

					// Clean up.
					for _, createdAccount := range actual.Accounts {
						assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
					}
				}
			})
		})

		test.Run("Reading", func(subtest *testing.T) {
			subtest.Parallel()

			runTestForAllAuthMethods(subtest, "it should return an error when trying to read something that does not exist", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Attempt to fetch nonexistent account.
					_, err := testClient.GetAccount(ctx, nonexistentID)
					assert.Error(t, err)
				}
			})

			runTestForAllAuthMethods(subtest, "it should be readable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create account.
					exampleAccount := fakes.BuildFakeAccount()
					exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
					createdAccount, err := testClient.CreateAccount(ctx, exampleAccountInput)
					checkValueAndError(t, createdAccount, err)

					// Fetch account.
					actual, err := testClient.GetAccount(ctx, createdAccount.ID)
					checkValueAndError(t, actual, err)

					// Assert account equality.
					checkAccountEquality(t, exampleAccount, actual)

					// Clean up account.
					assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
				}
			})
		})

		test.Run("Updating", func(subtest *testing.T) {
			subtest.Parallel()

			runTestForAllAuthMethods(subtest, "it should return an error when trying to update something that does not exist", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					exampleAccount := fakes.BuildFakeAccount()
					exampleAccount.ID = nonexistentID

					assert.Error(t, testClient.UpdateAccount(ctx, exampleAccount))
				}
			})

			runTestForAllAuthMethods(subtest, "it should be updateable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create account.
					exampleAccount := fakes.BuildFakeAccount()
					exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
					createdAccount, err := testClient.CreateAccount(ctx, exampleAccountInput)
					checkValueAndError(t, createdAccount, err)

					// Change account.
					createdAccount.Update(converters.ConvertAccountToAccountUpdateInput(exampleAccount))
					assert.NoError(t, testClient.UpdateAccount(ctx, createdAccount))

					// Fetch account.
					actual, err := testClient.GetAccount(ctx, createdAccount.ID)
					checkValueAndError(t, actual, err)

					// Assert account equality.
					checkAccountEquality(t, exampleAccount, actual)
					assert.NotNil(t, actual.LastUpdatedOn)

					adminClientLock.Lock()
					defer adminClientLock.Unlock()

					auditLogEntries, err := adminCookieClient.GetAuditLogForAccount(ctx, createdAccount.ID)
					require.NoError(t, err)

					expectedAuditLogEntries := []*types.AuditLogEntry{
						{EventType: audit.AccountCreationEvent},
						{EventType: audit.AccountUpdateEvent},
					}
					validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccount.ID, audit.AccountAssignmentKey)

					// Clean up account.
					assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
				}
			})
		})

		test.Run("Deleting", func(subtest *testing.T) {
			subtest.Parallel()

			runTestForAllAuthMethods(subtest, "it should return an error when trying to delete something that does not exist", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					assert.Error(t, testClient.ArchiveAccount(ctx, nonexistentID))
				}
			})

			runTestForAllAuthMethods(subtest, "it should be deletable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create account.
					exampleAccount := fakes.BuildFakeAccount()
					exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
					createdAccount, err := testClient.CreateAccount(ctx, exampleAccountInput)
					checkValueAndError(t, createdAccount, err)

					// Clean up account.
					assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))

					adminClientLock.Lock()
					defer adminClientLock.Unlock()

					auditLogEntries, err := adminCookieClient.GetAuditLogForAccount(ctx, createdAccount.ID)
					require.NoError(t, err)

					expectedAuditLogEntries := []*types.AuditLogEntry{
						{EventType: audit.AccountCreationEvent},
						{EventType: audit.AccountArchiveEvent},
					}
					validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAccount.ID, audit.AccountAssignmentKey)
				}
			})
		})
	*/

	test.Run("Memberships", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "should be creatable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// fetch account data
				accountCreationInput := &types.AccountCreationInput{
					Name:                   fakes.BuildFakeAccount().Name,
					DefaultUserPermissions: permissions.ServiceUserPermissions(math.MaxUint32),
				}
				account, accountCreationErr := testClient.CreateAccount(ctx, accountCreationInput)
				require.NoError(t, accountCreationErr)
				require.NotNil(t, account)

				userCount := 1

				require.Equal(t, accountCreationInput.DefaultUserPermissions, account.DefaultUserPermissions, "expected and actual permissions do not match")

				require.NoError(t, testClient.SetOption(httpclient.UsingAccount(account.ID)))

				t.Logf("created account #%d", account.ID)

				// create a webhook
				exampleWebhook := fakes.BuildFakeWebhook()
				exampleWebhookInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
				createdWebhook, creationErr := testClient.CreateWebhook(ctx, exampleWebhookInput)
				checkValueAndError(t, createdWebhook, creationErr)

				// create dummy users
				users := []*types.User{}
				clients := []*httpclient.Client{}

				for i := 0; i < userCount; i++ {
					u, _, _, c := createUserAndClientForTest(ctx, t)
					users = append(users, u)
					clients = append(clients, c)
				}

				// check that each user cannot see the webhook
				for i := 0; i < userCount; i++ {
					webhook, err := clients[i].GetWebhook(ctx, createdWebhook.ID)
					require.Nil(t, webhook)
					require.Error(t, err)
				}

				// add them to the account
				for i := 0; i < userCount; i++ {
					require.NoError(t, testClient.AddUserToAccount(ctx, account.ID, &types.AddUserToAccountInput{UserID: users[i].ID, Reason: t.Name()}))
					require.NoError(t, clients[i].SetOption(httpclient.UsingAccount(account.ID)))
				}

				// check that each user can see the webhook
				for i := 0; i < userCount; i++ {
					webhook, err := clients[i].GetWebhook(ctx, createdWebhook.ID)
					require.Nil(t, webhook)
					require.Error(t, err)
				}

				// reduce all permissions to nothing
				for i := 0; i < userCount; i++ {
					input := &types.ModifyUserPermissionsInput{
						UserID:                 users[i].ID,
						UserAccountPermissions: 0,
						Reason:                 t.Name(),
					}
					require.NoError(t, testClient.ModifyMemberPermissions(ctx, account.ID, input))
				}

				// check that each user cannot update the webhook
				for i := 0; i < userCount; i++ {
					require.Error(t, clients[i].UpdateWebhook(ctx, createdWebhook))
				}

				// restore all permissions
				for i := 0; i < userCount; i++ {
					input := &types.ModifyUserPermissionsInput{
						UserID:                 users[i].ID,
						UserAccountPermissions: testutil.BuildMaxUserPerms(),
						Reason:                 t.Name(),
					}
					require.NoError(t, testClient.ModifyMemberPermissions(ctx, account.ID, input))
				}

				// check that each user can update the webhooks
				for i := 0; i < userCount; i++ {
					require.NoError(t, clients[i].UpdateWebhook(ctx, createdWebhook))
				}

				// remove users from account
				for i := 0; i < userCount; i++ {
					require.NoError(t, testClient.RemoveUser(ctx, account.ID, users[i].ID))
				}

				// check that each user cannot see the webhooks
				for i := 0; i < userCount; i++ {
					webhook, err := clients[i].GetWebhook(ctx, createdWebhook.ID)
					require.Nil(t, webhook)
					require.Error(t, err)
				}

				// check audit entries
				adminClientLock.Lock()
				defer adminClientLock.Unlock()

				auditLogEntries, err := adminCookieClient.GetAuditLogForAccount(ctx, account.ID)
				require.NoError(t, err)

				expectedAuditLogEntries := []*types.AuditLogEntry{
					{EventType: audit.AccountCreationEvent},
					{EventType: audit.UserAddedToAccountEvent},
					{EventType: audit.WebhookCreationEvent},
					{EventType: audit.WebhookCreationEvent},
				}
				validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, account.ID, audit.AccountAssignmentKey)

				// Clean up.
				for i := 0; i < userCount; i++ {
					require.NoError(t, testClient.ArchiveUser(ctx, users[i].ID))
				}
			}
		})
	})

	/*
		test.Run("Auditing", func(subtest *testing.T) {
			subtest.Parallel()

			runTestForAllAuthMethods(subtest, "it should return an error when trying to audit something that does not exist", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					adminClientLock.Lock()
					defer adminClientLock.Unlock()
					x, err := adminCookieClient.GetAuditLogForAccount(ctx, nonexistentID)

					assert.NoError(t, err)
					assert.Empty(t, x)
				}
			})

			runTestForAllAuthMethods(subtest, "it should not be auditable by a non-admin", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create account.
					exampleAccount := fakes.BuildFakeAccount()
					exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
					createdAccount, err := testClient.CreateAccount(ctx, exampleAccountInput)
					checkValueAndError(t, createdAccount, err)

					// fetch audit log entries
					actual, err := testClient.GetAuditLogForAccount(ctx, createdAccount.ID)
					assert.Error(t, err)
					assert.Nil(t, actual)

					// Clean up account.
					assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
				}
			})
		})
	*/
}
