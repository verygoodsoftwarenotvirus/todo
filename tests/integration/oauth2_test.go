package integration

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func mustBuildCode(t *testing.T, totpSecret string) string {
	t.Helper()

	code, err := totp.GenerateCode(totpSecret, time.Now().UTC())
	require.NoError(t, err)

	return code
}

func buildDummyOAuth2ClientInput(t *testing.T, username, password, totpToken string) *types.OAuth2ClientCreationInput {
	t.Helper()

	x := &types.OAuth2ClientCreationInput{
		UserLoginInput: types.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: mustBuildCode(t, totpToken),
		},
		Scopes:      []string{"*"},
		RedirectURI: "http://localhost",
	}

	return x
}

func convertInputToClient(input *types.OAuth2ClientCreationInput) *types.OAuth2Client {
	return &types.OAuth2Client{
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
	}
}

func checkOAuth2ClientEquality(t *testing.T, expected, actual *types.OAuth2Client) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.NotEmpty(t, actual.ClientID)
	assert.NotEmpty(t, actual.ClientSecret)
	assert.Equal(t, expected.RedirectURI, actual.RedirectURI)
	assert.Equal(t, expected.Scopes, actual.Scopes)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func createOAuth2Client(ctx context.Context, t *testing.T, testUser *types.User, testClient *httpclient.V1Client) (*types.OAuth2ClientCreationInput, *types.OAuth2Client) {
	cookie, err := testClient.Login(ctx, &types.UserLoginInput{
		Username:  testUser.Username,
		Password:  testUser.HashedPassword,
		TOTPToken: generateTOTPTokenForUser(t, testUser),
	})
	require.NoError(t, err)

	clientInput := fakes.BuildFakeOAuth2ClientCreationInput()
	input := &types.OAuth2ClientCreationInput{
		UserLoginInput: types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		},
		Name:         clientInput.Name,
		ClientID:     clientInput.ClientID,
		ClientSecret: clientInput.ClientSecret,
		RedirectURI:  clientInput.RedirectURI,
		Scopes:       []string{"*"},
	}

	createdClient, err := testClient.CreateOAuth2Client(ctx, cookie, input)

	checkValueAndError(t, createdClient, err)
	checkOAuth2ClientEquality(t, convertInputToClient(input), createdClient)

	return input, createdClient
}

func TestOAuth2Clients(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be creatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)
			cookie, err := testClient.Login(ctx, &types.UserLoginInput{
				Username:  testUser.Username,
				Password:  testUser.HashedPassword,
				TOTPToken: generateTOTPTokenForUser(t, testUser),
			})
			require.NoError(t, err)

			// create oauth2 client
			clientInput := fakes.BuildFakeOAuth2ClientCreationInput()
			input := &types.OAuth2ClientCreationInput{
				UserLoginInput: types.UserLoginInput{
					Username:  testUser.Username,
					Password:  testUser.HashedPassword,
					TOTPToken: generateTOTPTokenForUser(t, testUser),
				},
				Name:         clientInput.Name,
				ClientID:     clientInput.ClientID,
				ClientSecret: clientInput.ClientSecret,
				RedirectURI:  clientInput.RedirectURI,
				Scopes:       []string{"*"},
			}

			createdClient, err := testClient.CreateOAuth2Client(ctx, cookie, input)

			checkValueAndError(t, createdClient, err)
			checkOAuth2ClientEquality(t, convertInputToClient(input), createdClient)

			// Clean up.
			assert.NoError(t, testClient.ArchiveOAuth2Client(ctx, createdClient.ID))
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to read one that doesn'subtest exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Fetch oauth2Client.
			_, err := testClient.GetOAuth2Client(ctx, nonexistentID)
			assert.Error(t, err)
		})

		subtest.Run("it should be readable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)

			// Create oauth2Client.
			input, createdClient := createOAuth2Client(ctx, t, testUser, testClient)

			// Fetch oauth2Client.
			actual, err := testClient.GetOAuth2Client(ctx, createdClient.ID)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality.
			checkOAuth2ClientEquality(t, convertInputToClient(input), actual)

			// Clean up.
			err = testClient.ArchiveOAuth2Client(ctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)

			// Create oauth2Client.
			_, createdClient := createOAuth2Client(ctx, t, testUser, testClient)

			// Clean up.
			assert.NoError(t, testClient.ArchiveOAuth2Client(ctx, createdClient.ID))
		})

		subtest.Run("should be unable to authorize after being deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)
			cookie, err := testClient.Login(ctx, &types.UserLoginInput{
				Username:  testUser.Username,
				Password:  testUser.HashedPassword,
				TOTPToken: generateTOTPTokenForUser(t, testUser),
			})
			require.NoError(t, err)

			input := buildDummyOAuth2ClientInput(test, testUser.Username, testUser.HashedPassword, testUser.TwoFactorSecret)
			premade, err := testClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(test, premade, err)

			// archive oauth2Client.
			require.NoError(t, testClient.ArchiveOAuth2Client(ctx, premade.ID))

			c2, err := httpclient.NewClient(
				ctx,
				premade.ClientID,
				premade.ClientSecret,
				testClient.URL,
				noop.NewLogger(),
				buildHTTPClient(),
				premade.Scopes,
				true,
			)
			checkValueAndError(test, c2, err)

			_, err = c2.GetOAuth2Clients(ctx, nil)
			assert.Error(t, err, "expected error from what should be an unauthorized client")
		})
	})

	test.Run("Listing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)

			// Create oauth2Clients.
			var expected []*types.OAuth2Client
			for i := 0; i < 5; i++ {
				// Create oauth2Client.
				_, oac := createOAuth2Client(ctx, t, testUser, testClient)
				expected = append(expected, oac)
			}

			// Assert oauth2Client list equality.
			actual, err := testClient.GetOAuth2Clients(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(actual.Clients)-len(expected) > 0,
				"expected %d - %d to be > 0",
				len(actual.Clients),
				len(expected),
			)

			for _, oAuth2Client := range expected {
				clientFound := false
				for _, c := range actual.Clients {
					if c.ID == oAuth2Client.ID {
						clientFound = true
						break
					}
				}
				assert.True(t, clientFound, "expected oAuth2Client ID %d to be present in results", oAuth2Client.ID)
			}

			// Clean up.
			for _, oa2c := range expected {
				err = testClient.ArchiveOAuth2Client(ctx, oa2c.ID)
				assert.NoError(t, err, "error deleting client %d: %v", oa2c.ID, err)
			}
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
			exampleOAuth2Client.ID = nonexistentID

			x, err := adminClient.GetAuditLogForOAuth2Client(ctx, exampleOAuth2Client.ID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		subtest.Run("it should be auditable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)

			// Create oauth2Client.
			_, createdClient := createOAuth2Client(ctx, t, testUser, testClient)

			// fetch audit log entries
			actual, err := adminClient.GetAuditLogForOAuth2Client(ctx, createdClient.ID)
			assert.NoError(t, err)
			assert.Len(t, actual, 1)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveOAuth2Client(ctx, createdClient.ID))
		})

		subtest.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			testUser, testClient := createUserAndClientForTest(ctx, t)

			// Create oauth2Client.
			_, createdClient := createOAuth2Client(ctx, t, testUser, testClient)

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForOAuth2Client(ctx, createdClient.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveOAuth2Client(ctx, createdClient.ID))
		})
	})
}
