package httpclient

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/http_client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/noop"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustBuildCode(t *testing.T, totpSecret string) string {
	t.Helper()
	code, err := totp.GenerateCode(totpSecret, time.Now())
	require.NoError(t, err)
	return code
}

func buildDummyOAuth2ClientInput(t *testing.T, username, password, TOTPToken string) *models.OAuth2ClientCreationInput {
	t.Helper()

	x := &models.OAuth2ClientCreationInput{
		UserLoginInput: models.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: mustBuildCode(t, TOTPToken),
		},
		Scopes:      []string{"*"},
		RedirectURI: localTestInstanceURL,
	}

	return x
}

func checkOAuth2ClientEquality(t *testing.T, expected *models.OAuth2ClientCreationInput, actual *models.OAuth2Client) {
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
	test.Parallel()

	// create user
	x, y, cookie := buildDummyUser(test)
	assert.NotNil(test, cookie)

	input := buildDummyOAuth2ClientInput(test, x.Username, y.Password, x.TwoFactorSecret)
	premade, err := todoClient.CreateOAuth2Client(context.Background(), input, cookie)
	checkValueAndError(test, premade, err)

	testClient, err := client.NewClient(
		premade.ClientID,
		premade.ClientSecret,
		todoClient.URL,
		noop.ProvideNoopLogger(),
		todoClient.PlainClient(),
		tracing.ProvideNoopTracer(),
		debug,
	)
	require.NoError(test, err, "error setting up auxiliary client")

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be creatable", func(t *testing.T) {
			tctx := context.Background()

			// Create oauth2Client
			actual, err := testClient.CreateOAuth2Client(tctx, input, cookie)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality
			checkOAuth2ClientEquality(t, input, actual)

			// Clean up
			err = testClient.DeleteOAuth2Client(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read one that doesn't exist", func(t *testing.T) {
			tctx := context.Background()

			// Fetch oauth2Client
			_, err := testClient.GetOAuth2Client(tctx, nonexistentID)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := context.Background()

			// Create oauth2Client
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
			c, err := testClient.CreateOAuth2Client(tctx, input, cookie)
			checkValueAndError(t, c, err)

			t.Logf("premade.ID:       %d\n", c.ID)
			t.Logf("premade.ClientID: %s\n", c.ClientID)

			// Fetch oauth2Client
			actual, err := testClient.GetOAuth2Client(tctx, c.ID)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality
			checkOAuth2ClientEquality(t, input, actual)

			// Clean up
			err = testClient.DeleteOAuth2Client(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := context.Background()

			// Create oauth2Client
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
			premade, err := testClient.CreateOAuth2Client(tctx, input, cookie)
			checkValueAndError(t, premade, nil)

			// Clean up
			err = testClient.DeleteOAuth2Client(tctx, premade.ID)
			assert.NoError(t, err)
		})

		T.Run("should be unable to authorize after being deleted", func(t *testing.T) {
			tctx := context.Background()

			// create user
			createdUser, createdUserInput, _ := buildDummyUser(test)
			assert.NotNil(test, cookie)

			input := buildDummyOAuth2ClientInput(test, createdUserInput.Username, createdUserInput.Password, createdUser.TwoFactorSecret)
			premade, err := todoClient.CreateOAuth2Client(context.Background(), input, cookie)
			checkValueAndError(test, premade, err)

			// Delete oauth2Client
			err = testClient.DeleteOAuth2Client(tctx, premade.ID)
			assert.NoError(t, err)

			c2, err := client.NewClient(
				premade.ClientID,
				premade.ClientSecret,
				todoClient.URL,
				noop.ProvideNoopLogger(),
				buildHTTPClient(),
				tracing.ProvideNoopTracer(),
				true,
			)
			checkValueAndError(test, c2, err)

			_, err = c2.GetItems(tctx, nil)
			t.Logf("%v", err)
			assert.Error(t, err, "expected error from what should be an unauthorized client")
			// assert.Equal(t, "", err.Error())
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			tctx := context.Background()

			// Create oauth2Clients
			var expected []*models.OAuth2Client
			for i := 0; i < 5; i++ {
				input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
				oac, err := testClient.CreateOAuth2Client(tctx, input, cookie)
				checkValueAndError(t, oac, err)
				expected = append(expected, oac)
			}

			// Assert oauth2Client list equality
			actual, err := testClient.GetOAuth2Clients(tctx, nil)
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

			// Clean up
			for _, oa2c := range expected {
				err = testClient.DeleteOAuth2Client(tctx, oa2c.ID)
				assert.NoError(t, err, "error deleting client %d: %v", oa2c.ID, err)
			}
		})
	})

}