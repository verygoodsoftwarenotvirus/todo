package integration

import (
	"context"
	"testing"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

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

func buildDummyOAuth2ClientInput(t *testing.T, username, password, totpToken string) *models.OAuth2ClientCreationInput {
	t.Helper()

	x := &models.OAuth2ClientCreationInput{
		UserLoginInput: models.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: mustBuildCode(t, totpToken),
		},
		Scopes:      []string{"*"},
		RedirectURI: "http://localhost",
	}

	return x
}

func convertInputToClient(input *models.OAuth2ClientCreationInput) *models.OAuth2Client {
	return &models.OAuth2Client{
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		RedirectURI:   input.RedirectURI,
		Scopes:        input.Scopes,
		BelongsToUser: input.BelongsToUser,
	}
}

func checkOAuth2ClientEquality(t *testing.T, expected, actual *models.OAuth2Client) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.NotEmpty(t, actual.ClientID)
	assert.NotEmpty(t, actual.ClientSecret)
	assert.Equal(t, expected.RedirectURI, actual.RedirectURI)
	assert.Equal(t, expected.Scopes, actual.Scopes)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func TestOAuth2Clients(test *testing.T) {
	_ctx := context.Background()

	// create user.
	x, y, cookie := buildDummyUser(_ctx, test)
	assert.NotNil(test, cookie)

	twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(x.TwoFactorQRCode)
	require.NoError(test, err)

	input := buildDummyOAuth2ClientInput(test, x.Username, y.Password, twoFactorSecret)
	premade, err := todoClient.CreateOAuth2Client(_ctx, cookie, input)
	checkValueAndError(test, premade, err)

	testClient, err := client.NewClient(
		_ctx,
		premade.ClientID,
		premade.ClientSecret,
		todoClient.URL,
		noop.NewLogger(),
		todoClient.PlainClient(),
		premade.Scopes,
		debug,
	)
	require.NoError(test, err, "error setting up auxiliary client")

	test.Run("Creating", func(t *testing.T) {
		t.Run("should be creatable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create oauth2Client.
			actual, err := testClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality.
			checkOAuth2ClientEquality(t, convertInputToClient(input), actual)

			// Clean up.
			err = testClient.ArchiveOAuth2Client(ctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Reading", func(t *testing.T) {
		t.Run("it should return an error when trying to read one that doesn't exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Fetch oauth2Client.
			_, err := testClient.GetOAuth2Client(ctx, nonexistentID)
			assert.Error(t, err)
		})

		t.Run("it should be readable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create oauth2Client.
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, twoFactorSecret)
			c, err := testClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(t, c, err)

			// Fetch oauth2Client.
			actual, err := testClient.GetOAuth2Client(ctx, c.ID)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality.
			checkOAuth2ClientEquality(t, convertInputToClient(input), actual)

			// Clean up.
			err = testClient.ArchiveOAuth2Client(ctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(t *testing.T) {
		t.Run("should be able to be deleted", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create oauth2Client.
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, twoFactorSecret)
			premade, err := testClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(t, premade, err)

			// Clean up.
			err = testClient.ArchiveOAuth2Client(ctx, premade.ID)
			assert.NoError(t, err)
		})

		t.Run("should be unable to authorize after being deleted", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// create user.
			createdUser, createdUserInput, _ := buildDummyUser(ctx, test)
			assert.NotNil(test, cookie)

			createdUserTwoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(createdUser.TwoFactorQRCode)
			require.NoError(t, err)

			input := buildDummyOAuth2ClientInput(test, createdUserInput.Username, createdUserInput.Password, createdUserTwoFactorSecret)
			premade, err := todoClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(test, premade, err)

			// archive oauth2Client.
			require.NoError(t, testClient.ArchiveOAuth2Client(ctx, premade.ID))

			c2, err := client.NewClient(
				ctx,
				premade.ClientID,
				premade.ClientSecret,
				todoClient.URL,
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

	test.Run("Listing", func(t *testing.T) {
		t.Run("should be able to be read in a list", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create oauth2Clients.
			var expected []*models.OAuth2Client
			for i := 0; i < 5; i++ {
				input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, twoFactorSecret)
				oac, err := testClient.CreateOAuth2Client(ctx, cookie, input)
				checkValueAndError(t, oac, err)
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

	test.Run("Auditing", func(t *testing.T) {
		t.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()
			exampleOAuth2Client.ID = nonexistentID

			x, err := adminClient.GetAuditLogForOAuth2Client(ctx, exampleOAuth2Client.ID)
			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		t.Run("it should be auditable", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create oauth2Client.
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, twoFactorSecret)
			createdOAuth2Client, err := testClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(t, createdOAuth2Client, err)

			// fetch audit log entries
			actual, err := adminClient.GetAuditLogForOAuth2Client(ctx, createdOAuth2Client.ID)
			assert.NoError(t, err)
			assert.Len(t, actual, 1)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveOAuth2Client(ctx, createdOAuth2Client.ID))
		})

		t.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			// Create oauth2Client.
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, twoFactorSecret)
			createdOAuth2Client, err := testClient.CreateOAuth2Client(ctx, cookie, input)
			checkValueAndError(t, createdOAuth2Client, err)

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForOAuth2Client(ctx, createdOAuth2Client.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveOAuth2Client(ctx, createdOAuth2Client.ID))
		})
	})
}
