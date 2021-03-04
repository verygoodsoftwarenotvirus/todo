package integration

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkAPIClientEquality(t *testing.T, expected, actual *types.APIClient) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for API client #%d to be %q, but it was %q ", actual.ID, expected.Name, actual.Name)
	assert.NotEmpty(t, actual.ExternalID, "expected ExternalID for API client #%d to not be empty, but it was", actual.ID)
	assert.NotEmpty(t, actual.ClientID, "expected ClientID for API client #%d to not be empty, but it was", actual.ID)
	assert.Empty(t, actual.ClientSecret, "expected ClientSecret for API client #%d to not be empty, but it was", actual.ID)
	assert.NotZero(t, actual.BelongsToAccount, "expected BelongsToAccount for API client #%d to not be zero, but it was", actual.ID)
	assert.NotZero(t, actual.CreatedOn)
}

func TestAPIClients(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "should be creatable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create API client.
				exampleAPIClient := fakes.BuildFakeAPIClient()
				exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
				exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdAPIClient, err := testClient.CreateAPIClient(ctx, cookie, exampleAPIClientInput)
				checkValueAndError(t, createdAPIClient, err)

				// Assert API client equality.
				assert.NotEmpty(t, createdAPIClient.ClientID, "expected ClientID for API client #%d to not be empty, but it was", createdAPIClient.ID)
				assert.NotEmpty(t, createdAPIClient.ClientSecret, "expected ClientSecret for API client #%d to not be empty, but it was", createdAPIClient.ID)

				adminClientLock.Lock()
				defer adminClientLock.Unlock()

				auditLogEntries, err := adminCookieClient.GetAuditLogForAPIClient(ctx, createdAPIClient.ID)
				require.NoError(t, err)

				expectedAuditLogEntries := []*types.AuditLogEntry{
					{EventType: audit.APIClientCreationEvent},
				}
				validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAPIClient.ID, audit.APIClientAssignmentKey)

				// Clean up.
				assert.NoError(t, testClient.ArchiveAPIClient(ctx, createdAPIClient.ID))
			}
		})
	})

	test.Run("Listing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			ctx := context.Background()
			user, cookie, testClient, _ := createUserAndClientForTest(ctx, t)

			// Create API clients.
			var expected []uint64
			for i := 0; i < 5; i++ {
				// Create API client.
				exampleAPIClient := fakes.BuildFakeAPIClient()
				exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
				exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}
				createdAPIClient, apiClientCreationErr := testClient.CreateAPIClient(ctx, cookie, exampleAPIClientInput)
				checkValueAndError(t, createdAPIClient, apiClientCreationErr)

				expected = append(expected, createdAPIClient.ID)
			}

			// Assert API client list equality.
			actual, err := testClient.GetAPIClients(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Clients),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Clients),
			)

			// Clean up.
			for _, createdAPIClient := range actual.Clients {
				assert.NoError(t, testClient.ArchiveAPIClient(ctx, createdAPIClient.ID))
			}
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "it should return an error when trying to read something that does not exist", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Attempt to fetch nonexistent API client.
				_, err := testClient.GetAPIClient(ctx, nonexistentID)
				assert.Error(t, err)
			}
		})

		runTestForAllAuthMethods(subtest, "it should be readable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create API client.
				exampleAPIClient := fakes.BuildFakeAPIClient()
				exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
				exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdAPIClient, err := testClient.CreateAPIClient(ctx, cookie, exampleAPIClientInput)
				checkValueAndError(t, createdAPIClient, err)

				// Fetch API client.
				actual, err := testClient.GetAPIClient(ctx, createdAPIClient.ID)
				checkValueAndError(t, actual, err)

				// Assert API client equality.
				checkAPIClientEquality(t, exampleAPIClient, actual)

				// Clean up API client.
				assert.NoError(t, testClient.ArchiveAPIClient(ctx, createdAPIClient.ID))
			}
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "it should return an error when trying to delete something that does not exist", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				assert.Error(t, testClient.ArchiveAPIClient(ctx, nonexistentID))
			}
		})

		runTestForAllAuthMethods(subtest, "it should be deletable", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create API client.
				exampleAPIClient := fakes.BuildFakeAPIClient()
				exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
				exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdAPIClient, err := testClient.CreateAPIClient(ctx, cookie, exampleAPIClientInput)
				checkValueAndError(t, createdAPIClient, err)

				// Clean up API client.
				assert.NoError(t, testClient.ArchiveAPIClient(ctx, createdAPIClient.ID))

				adminClientLock.Lock()
				defer adminClientLock.Unlock()

				auditLogEntries, err := adminCookieClient.GetAuditLogForAPIClient(ctx, createdAPIClient.ID)
				require.NoError(t, err)

				expectedAuditLogEntries := []*types.AuditLogEntry{
					{EventType: audit.APIClientCreationEvent},
					{EventType: audit.APIClientArchiveEvent},
				}
				validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAPIClient.ID, audit.APIClientAssignmentKey)
			}
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		runTestForAllAuthMethods(subtest, "it should not be auditable by a non-admin", func(ctx context.Context, user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create API client.
				exampleAPIClient := fakes.BuildFakeAPIClient()
				exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
				exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdAPIClient, err := testClient.CreateAPIClient(ctx, cookie, exampleAPIClientInput)
				checkValueAndError(t, createdAPIClient, err)

				// fetch audit log entries
				actual, err := testClient.GetAuditLogForAPIClient(ctx, createdAPIClient.ID)
				assert.Error(t, err)
				assert.Nil(t, actual)

				// Clean up API client.
				assert.NoError(t, testClient.ArchiveAPIClient(ctx, createdAPIClient.ID))
			}
		})
	})
}
