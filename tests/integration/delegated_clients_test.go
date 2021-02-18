package integration

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkDelegatedClientEquality(t *testing.T, expected, actual *types.DelegatedClient) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for delegated client #%d to be %q, but it was %q ", actual.ID, expected.Name, actual.Name)
	assert.NotEmpty(t, actual.ExternalID, "expected ExternalID for delegated client #%d to not be empty, but it was", actual.ID)
	assert.NotEmpty(t, actual.ClientID, "expected ClientID for delegated client #%d to not be empty, but it was", actual.ID)
	assert.Empty(t, actual.ClientSecret, "expected ClientSecret for delegated client #%d to not be empty, but it was", actual.ID)
	assert.NotZero(t, actual.BelongsToUser, "expected BelongsToUser for delegated client #%d to not be zero, but it was", actual.ID)
	assert.NotZero(t, actual.CreatedOn)
}

func TestDelegatedClients(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		runTestForAllAuthMethods(ctx, subtest, "should be creatable", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create delegated client.
				exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
				exampleDelegatedClientInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)
				exampleDelegatedClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				testClient.SetOption(httpclient.WithDebug())

				createdDelegatedClient, err := testClient.CreateDelegatedClient(ctx, cookie, exampleDelegatedClientInput)
				checkValueAndError(t, createdDelegatedClient, err)

				// Assert delegated client equality.
				assert.NotEmpty(t, createdDelegatedClient.ClientID, "expected ClientID for delegated client #%d to not be empty, but it was", createdDelegatedClient.ID)
				assert.NotEmpty(t, createdDelegatedClient.ClientSecret, "expected ClientSecret for delegated client #%d to not be empty, but it was", createdDelegatedClient.ID)

				adminClientLock.Lock()
				defer adminClientLock.Unlock()

				auditLogEntries, err := adminClient.GetAuditLogForDelegatedClient(ctx, createdDelegatedClient.ID)
				require.NoError(t, err)

				expectedAuditLogEntries := []*types.AuditLogEntry{
					{EventType: audit.DelegatedClientCreationEvent},
				}
				validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdDelegatedClient.ID, audit.DelegatedClientAssignmentKey)

				// Clean up.
				assert.NoError(t, testClient.ArchiveDelegatedClient(ctx, createdDelegatedClient.ID))
			}
		})
	})

	test.Run("Listing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			runTestForAllAuthMethods(ctx, subtest, "should be able to be read in a list", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
				return func(t *testing.T) {
					// Create delegated clients.
					var expected []uint64
					for i := 0; i < 5; i++ {
						// Create delegated client.
						exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
						exampleDelegatedClientInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)
						exampleDelegatedClientInput.UserLoginInput = types.UserLoginInput{
							Username:  user.Username,
							Password:  user.HashedPassword,
							TOTPToken: generateTOTPTokenForUser(t, user),
						}
						createdDelegatedClient, delegatedClientCreationErr := testClient.CreateDelegatedClient(ctx, cookie, exampleDelegatedClientInput)
						checkValueAndError(t, createdDelegatedClient, delegatedClientCreationErr)

						expected = append(expected, createdDelegatedClient.ID)
					}

					// Assert delegated client list equality.
					actual, err := testClient.GetDelegatedClients(ctx, nil)
					checkValueAndError(t, actual, err)
					assert.True(
						t,
						len(expected) <= len(actual.Clients),
						"expected %d to be <= %d",
						len(expected),
						len(actual.Clients),
					)

					// Clean up.
					for _, createdDelegatedClient := range actual.Clients {
						assert.NoError(t, testClient.ArchiveDelegatedClient(ctx, createdDelegatedClient.ID))
					}
				}
			})
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		runTestForAllAuthMethods(ctx, subtest, "it should return an error when trying to read something that does not exist", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Attempt to fetch nonexistent delegated client.
				_, err := testClient.GetDelegatedClient(ctx, nonexistentID)
				assert.Error(t, err)
			}
		})

		runTestForAllAuthMethods(ctx, subtest, "it should be readable", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create delegated client.
				exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
				exampleDelegatedClientInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)
				exampleDelegatedClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdDelegatedClient, err := testClient.CreateDelegatedClient(ctx, cookie, exampleDelegatedClientInput)
				checkValueAndError(t, createdDelegatedClient, err)

				// Fetch delegated client.
				actual, err := testClient.GetDelegatedClient(ctx, createdDelegatedClient.ID)
				checkValueAndError(t, actual, err)

				// Assert delegated client equality.
				checkDelegatedClientEquality(t, exampleDelegatedClient, actual)

				// Clean up delegated client.
				assert.NoError(t, testClient.ArchiveDelegatedClient(ctx, createdDelegatedClient.ID))
			}
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		runTestForAllAuthMethods(ctx, subtest, "it should return an error when trying to delete something that does not exist", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				assert.Error(t, testClient.ArchiveDelegatedClient(ctx, nonexistentID))
			}
		})

		runTestForAllAuthMethods(ctx, subtest, "it should be deletable", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create delegated client.
				exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
				exampleDelegatedClientInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)
				exampleDelegatedClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdDelegatedClient, err := testClient.CreateDelegatedClient(ctx, cookie, exampleDelegatedClientInput)
				checkValueAndError(t, createdDelegatedClient, err)

				// Clean up delegated client.
				assert.NoError(t, testClient.ArchiveDelegatedClient(ctx, createdDelegatedClient.ID))

				adminClientLock.Lock()
				defer adminClientLock.Unlock()

				auditLogEntries, err := adminClient.GetAuditLogForDelegatedClient(ctx, createdDelegatedClient.ID)
				require.NoError(t, err)

				expectedAuditLogEntries := []*types.AuditLogEntry{
					{EventType: audit.DelegatedClientCreationEvent},
					{EventType: audit.DelegatedClientArchiveEvent},
				}
				validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdDelegatedClient.ID, audit.DelegatedClientAssignmentKey)
			}
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		runTestForAllAuthMethods(ctx, subtest, "it should not be auditable by a non-admin", func(user *types.User, cookie *http.Cookie, testClient *httpclient.Client) func(*testing.T) {
			return func(t *testing.T) {
				// Create delegated client.
				exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
				exampleDelegatedClientInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)
				exampleDelegatedClientInput.UserLoginInput = types.UserLoginInput{
					Username:  user.Username,
					Password:  user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, user),
				}

				createdDelegatedClient, err := testClient.CreateDelegatedClient(ctx, cookie, exampleDelegatedClientInput)
				checkValueAndError(t, createdDelegatedClient, err)

				// fetch audit log entries
				actual, err := testClient.GetAuditLogForDelegatedClient(ctx, createdDelegatedClient.ID)
				assert.Error(t, err)
				assert.Nil(t, actual)

				// Clean up delegated client.
				assert.NoError(t, testClient.ArchiveDelegatedClient(ctx, createdDelegatedClient.ID))
			}
		})
	})
}
