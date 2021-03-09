package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
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

func (s *TestSuite) TestAPIClientsCreating() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to create API clients via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create API client.
			exampleAPIClient := fakes.BuildFakeAPIClient()
			exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
			exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
				Username:  s.user.Username,
				Password:  s.user.HashedPassword,
				TOTPToken: generateTOTPTokenForUser(t, s.user),
			}

			createdAPIClient, err := testClients.main.CreateAPIClient(ctx, s.cookie, exampleAPIClientInput)
			requireNotNilAndNoProblems(t, createdAPIClient, err)

			// Assert API client equality.
			assert.NotEmpty(t, createdAPIClient.ClientID, "expected ClientID for API client #%d to not be empty, but it was", createdAPIClient.ID)
			assert.NotEmpty(t, createdAPIClient.ClientSecret, "expected ClientSecret for API client #%d to not be empty, but it was", createdAPIClient.ID)

			auditLogEntries, err := testClients.admin.GetAuditLogForAPIClient(ctx, createdAPIClient.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.APIClientCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAPIClient.ID, audit.APIClientAssignmentKey)

			// Clean up.
			assert.NoError(t, testClients.main.ArchiveAPIClient(ctx, createdAPIClient.ID))
		})
	}
}

func (s *TestSuite) TestAPIClientsListing() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to read API clients in a list via %s", authType), func() {
			t := s.T()

			const clientsToMake = 1

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create API clients.
			var expected []uint64
			for i := 0; i < clientsToMake; i++ {
				// Create API client.
				exampleAPIClient := fakes.BuildFakeAPIClient()
				exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
				exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
					Username:  s.user.Username,
					Password:  s.user.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, s.user),
				}
				createdAPIClient, apiClientCreationErr := testClients.main.CreateAPIClient(ctx, s.cookie, exampleAPIClientInput)
				requireNotNilAndNoProblems(t, createdAPIClient, apiClientCreationErr)

				expected = append(expected, createdAPIClient.ID)
			}

			// Assert API client list equality.
			actual, err := testClients.main.GetAPIClients(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Clients),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Clients),
			)

			// Clean up.
			for _, createdAPIClientID := range expected {
				assert.NoError(t, testClients.main.ArchiveAPIClient(ctx, createdAPIClientID))
			}
		})
	}
}

func (s *TestSuite) TestAPIClientsReading() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to read non-existent API clients via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Attempt to fetch nonexistent API client.
			_, err := testClients.main.GetAPIClient(ctx, nonexistentID)
			assert.Error(t, err)
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to read API clients via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create API client.
			exampleAPIClient := fakes.BuildFakeAPIClient()
			exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
			exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
				Username:  s.user.Username,
				Password:  s.user.HashedPassword,
				TOTPToken: generateTOTPTokenForUser(t, s.user),
			}

			createdAPIClient, err := testClients.main.CreateAPIClient(ctx, s.cookie, exampleAPIClientInput)
			requireNotNilAndNoProblems(t, createdAPIClient, err)

			// Fetch API client.
			actual, err := testClients.main.GetAPIClient(ctx, createdAPIClient.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert API client equality.
			checkAPIClientEquality(t, exampleAPIClient, actual)

			// Clean up API client.
			assert.NoError(t, testClients.main.ArchiveAPIClient(ctx, createdAPIClient.ID))
		})
	}
}

func (s *TestSuite) TestAPIClientsArchiving() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should not be possible to archive non-existent API clients via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
			defer span.End()

			assert.Error(t, testClients.main.ArchiveAPIClient(ctx, nonexistentID))
		})
	}

	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to archive API clients via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create API client.
			exampleAPIClient := fakes.BuildFakeAPIClient()
			exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
			exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
				Username:  s.user.Username,
				Password:  s.user.HashedPassword,
				TOTPToken: generateTOTPTokenForUser(t, s.user),
			}

			createdAPIClient, err := testClients.main.CreateAPIClient(ctx, s.cookie, exampleAPIClientInput)
			requireNotNilAndNoProblems(t, createdAPIClient, err)

			// Clean up API client.
			assert.NoError(t, testClients.main.ArchiveAPIClient(ctx, createdAPIClient.ID))

			auditLogEntries, err := testClients.admin.GetAuditLogForAPIClient(ctx, createdAPIClient.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.APIClientCreationEvent},
				{EventType: audit.APIClientArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdAPIClient.ID, audit.APIClientAssignmentKey)
		})
	}
}

func (s *TestSuite) TestAPIClientsAuditing() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be possible to audit API clients via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create API client.
			exampleAPIClient := fakes.BuildFakeAPIClient()
			exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
			exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
				Username:  s.user.Username,
				Password:  s.user.HashedPassword,
				TOTPToken: generateTOTPTokenForUser(t, s.user),
			}

			createdAPIClient, err := testClients.main.CreateAPIClient(ctx, s.cookie, exampleAPIClientInput)
			requireNotNilAndNoProblems(t, createdAPIClient, err)

			// fetch audit log entries
			actual, err := testClients.admin.GetAuditLogForAPIClient(ctx, createdAPIClient.ID)
			assert.NoError(t, err)
			assert.NotNil(t, actual)

			// Clean up API client.
			assert.NoError(t, testClients.main.ArchiveAPIClient(ctx, createdAPIClient.ID))
		})
	}
}
