package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkAccountEquality(t *testing.T, expected, actual *types.Account) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected BucketName for account #%d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.NotZero(t, actual.CreatedOn)
}

func TestAccounts(test *testing.T) {
	test.Parallel()

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

	test.Run("Searching", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "should be able to be search for accounts", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create accounts.
				exampleAccount := fakes.BuildFakeAccount()
				var expected []*types.Account
				for i := 0; i < 5; i++ {
					// Create account.
					exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
					exampleAccountInput.Name = fmt.Sprintf("%s %d", exampleAccountInput.Name, i)
					createdAccount, accountCreationErr := testClient.CreateAccount(ctx, exampleAccountInput)
					checkValueAndError(t, createdAccount, accountCreationErr)

					expected = append(expected, createdAccount)
				}

				exampleLimit := uint8(20)

				// Assert account list equality.
				actual, err := testClient.SearchAccounts(ctx, exampleAccount.Name, exampleLimit)
				checkValueAndError(t, actual, err)
				assert.True(
					t,
					len(expected) <= len(actual),
					"expected results length %d to be <= %d",
					len(expected),
					len(actual),
				)

				// Clean up.
				for _, createdAccount := range expected {
					assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
				}
			}
		})

		runTestForAllAuthMethods(subtest, "should only receive your own accounts", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				exampleLimit := uint8(20)
				_, _, clientA, _ := createUserAndClientForTest(ctx, t)
				_, _, clientB, _ := createUserAndClientForTest(ctx, t)

				// Create accounts for user A.
				exampleAccountA := fakes.BuildFakeAccount()
				var createdForA []*types.Account
				for i := 0; i < 5; i++ {
					// Create account.
					exampleAccountInputA := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccountA)
					exampleAccountInputA.Name = fmt.Sprintf("%s %d", exampleAccountInputA.Name, i)

					createdAccount, accountCreationErr := clientA.CreateAccount(ctx, exampleAccountInputA)
					checkValueAndError(t, createdAccount, accountCreationErr)

					createdForA = append(createdForA, createdAccount)
				}
				query := exampleAccountA.Name

				// Create accounts for user B.
				exampleAccountB := fakes.BuildFakeAccount()
				exampleAccountB.Name = reverse(exampleAccountA.Name)
				var createdForB []*types.Account
				for i := 0; i < 5; i++ {
					// Create account.
					exampleAccountInputB := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccountB)
					exampleAccountInputB.Name = fmt.Sprintf("%s %d", exampleAccountInputB.Name, i)

					createdAccount, accountCreationErr := clientB.CreateAccount(ctx, exampleAccountInputB)
					checkValueAndError(t, createdAccount, accountCreationErr)

					createdForB = append(createdForB, createdAccount)
				}

				expected := createdForA

				// Assert account list equality.
				actual, err := clientA.SearchAccounts(ctx, query, exampleLimit)
				checkValueAndError(t, actual, err)
				assert.True(
					t,
					len(expected) <= len(actual),
					"expected results length %d to be <= %d",
					len(expected),
					len(actual),
				)

				// Clean up.
				for _, createdAccount := range createdForA {
					assert.NoError(t, clientA.ArchiveAccount(ctx, createdAccount.ID))
				}

				for _, createdAccount := range createdForB {
					assert.NoError(t, clientB.ArchiveAccount(ctx, createdAccount.ID))
				}
			}
		})
	})

	test.Run("ExistenceChecking", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "should be able to be search for accounts", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Attempt to fetch nonexistent account.
				actual, err := testClient.AccountExists(ctx, nonexistentID)
				assert.NoError(t, err)
				assert.False(t, actual)
			}
		})

		runTestForAllAuthMethods(subtest, "it should return true with no error when the relevant account exists", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create account.
				exampleAccount := fakes.BuildFakeAccount()
				exampleAccountInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)
				createdAccount, err := testClient.CreateAccount(ctx, exampleAccountInput)
				checkValueAndError(t, createdAccount, err)

				// Fetch account.
				actual, err := testClient.AccountExists(ctx, createdAccount.ID)
				assert.NoError(t, err)
				assert.True(t, actual)

				// Clean up account.
				assert.NoError(t, testClient.ArchiveAccount(ctx, createdAccount.ID))
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
}
