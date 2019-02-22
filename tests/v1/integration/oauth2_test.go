package integration

import (
	"context"
	"strconv"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
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

func buildDummyOAuth2ClientInput(t *testing.T, username, password, totpSecret string) *models.OAuth2ClientCreationInput {
	t.Helper()

	x := &models.OAuth2ClientCreationInput{
		UserLoginInput: models.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: mustBuildCode(t, totpSecret),
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
	x, y, cookie := buildDummyUser(context.Background(), test)
	assert.NotNil(test, cookie)

	input := buildDummyOAuth2ClientInput(test, x.Username, y.Password, x.TwoFactorSecret)
	premade, err := todoClient.CreateOAuth2Client(context.Background(), input, cookie)
	checkValueAndError(test, premade, err)

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

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be creatable", func(t *testing.T) {
			tctx := buildSpanContext("create-oauth2-client")

			// Create oauth2Client
			actual, err := c2.CreateOAuth2Client(tctx, input, cookie)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality
			checkOAuth2ClientEquality(t, input, actual)

			// Clean up
			err = todoClient.DeleteOAuth2Client(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read one that doesn't exist", func(t *testing.T) {
			tctx := buildSpanContext("try-to-read-nonexistent-oauth2-client")

			// Fetch oauth2Client
			_, err := todoClient.GetOAuth2Client(tctx, strconv.Itoa(nonexistentID))
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := buildSpanContext("read-oauth2-client")

			// Create oauth2Client
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
			premade, err := c2.CreateOAuth2Client(tctx, input, cookie)
			checkValueAndError(t, premade, err)

			// Fetch oauth2Client
			actual, err := c2.GetOAuth2Client(tctx, premade.ClientID)
			checkValueAndError(t, actual, err)

			// Assert oauth2Client equality
			checkOAuth2ClientEquality(t, input, actual)

			// Clean up
			err = c2.DeleteOAuth2Client(tctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should be updatable", func(t *testing.T) {
			// tctx := buildSpanContext("CHANGEME")

			t.SkipNow()
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := buildSpanContext("delete-oauth2-client")

			// Create oauth2Client
			input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
			premade, err := c2.CreateOAuth2Client(tctx, input, cookie)
			checkValueAndError(t, premade, nil)

			// Clean up
			err = c2.DeleteOAuth2Client(tctx, premade.ID)
			assert.NoError(t, err)
		})

		T.Run("should be unable to authorize after being deleted", func(t *testing.T) {
			// tctx := buildSpanContext()

			t.SkipNow()
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			tctx := buildSpanContext("list-oauth2-clients")

			// Create oauth2Client
			// Create oauth2Clients
			var expected []*models.OAuth2Client
			for i := 0; i < 5; i++ {
				input := buildDummyOAuth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
				oac, err := c2.CreateOAuth2Client(tctx, input, cookie)
				checkValueAndError(t, oac, err)
				expected = append(expected, oac)
			}

			// Assert oauth2Client list equality
			actual, err := c2.GetOAuth2Clients(tctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Clients))

			// Clean up
			for _, oauth2Client := range actual.Clients {
				err = c2.DeleteOAuth2Client(tctx, oauth2Client.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Counting", func(T *testing.T) {
		T.Run("it should be able to be counted", func(t *testing.T) {
			// tctx := buildSpanContext("CHANGEME")

			t.SkipNow()
		})
	})

	test.Run("Using", func(T *testing.T) {
		T.Run("should allow an authorized client to use the implicit grant type", func(t *testing.T) {
			// tctx := buildSpanContext("CHANGEME")

			t.SkipNow()
		})

		T.Run("should not allow an unauthorized client to use the implicit grant type", func(t *testing.T) {
			// tctx := buildSpanContext("CHANGEME")

			t.SkipNow()
		})
	})
}
