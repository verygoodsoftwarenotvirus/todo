package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestAuth(test *testing.T) {
	test.Run("should be able to login and log out", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)
		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})

		assert.NotNil(t, cookie)
		assert.NoError(t, err)

		assert.Equal(t, authservice.CookieName, cookie.Name)
		assert.NotEmpty(t, cookie.Value)
		assert.NotZero(t, cookie.MaxAge)
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)

		assert.NoError(t, testClient.Logout(ctx))
	})

	test.Run("login request without body fails", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		_, testClient := createUserAndClientForTest(ctx, t)

		u, err := url.Parse(testClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
		checkValueAndError(t, req, err)

		// execute login request.
		res, err := testClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	test.Run("should not be able to log in with the wrong password", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)

		// create login request.
		var badPassword string
		for _, v := range testUser.HashedPassword {
			badPassword = string(v) + badPassword
		}

		r := &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  badPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		}

		cookie, err := testClient.Login(ctx, r)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	test.Run("should not be able to login as someone that doesn't exist", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)

		exampleUserCreationInput := fakes.BuildFakeUserCreationInput()
		r := &types.UserLoginInput{
			Username:  exampleUserCreationInput.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: "123456",
		}

		cookie, err := testClient.Login(ctx, r)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	test.Run("should not be able to login without validating TOTP secret", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient(ctx, t)

		// create a user.
		exampleUser := fakes.BuildFakeUser()
		exampleUserCreationInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
		ucr, err := testClient.CreateUser(ctx, exampleUserCreationInput)
		checkValueAndError(t, ucr, err)

		twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(ucr.TwoFactorQRCode)
		require.NoError(t, err)

		// create login request.
		token, err := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		r := &types.UserLoginInput{
			Username:  exampleUserCreationInput.Username,
			Password:  exampleUserCreationInput.Password,
			TOTPToken: token,
		}

		cookie, err := testClient.Login(ctx, r)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	test.Run("should reject an unauthenticated request", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient(ctx, t)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testClient.BuildURL(nil, "webhooks"), nil)
		assert.NoError(t, err)

		res, err := testClient.PlainClient().Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	test.Run("should be able to change password", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)

		// get login cookie
		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})
		require.NotNil(t, cookie)
		assert.NoError(t, err)

		// create new password.
		var backwardsPass string
		for _, v := range testUser.HashedPassword {
			backwardsPass = string(v) + backwardsPass
		}

		// create password update request.
		r := &types.PasswordUpdateInput{
			CurrentPassword: testUser.HashedPassword,
			TOTPToken:       generateTOTPTokenForUser(t, testUser),
			NewPassword:     backwardsPass,
		}
		out, err := json.Marshal(r)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		// TODO: BUILD REAL CHANGE PASSWORD METHOD IN HTTPCLIENT

		u, err := url.Parse(testClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/password/new"

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), body)
		checkValueAndError(t, req, err)
		req.AddCookie(cookie)

		// execute password update request.
		res, err := testClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "/auth/login", res.Request.URL.Path)

		// logout.

		u2, err := url.Parse(testClient.BuildURL(nil))
		require.NoError(t, err)
		u2.Path = "/users/logout"

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, u2.String(), nil)
		checkValueAndError(t, req, err)
		req.AddCookie(cookie)

		res, err = testClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// create login request.
		l, err := json.Marshal(&types.UserLoginInput{
			Username:  testUser.Username,
			Password:  backwardsPass,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})
		require.NoError(t, err)
		body = bytes.NewReader(l)

		u3, err := url.Parse(testClient.BuildURL(nil))
		require.NoError(t, err)
		u3.Path = "/users/login"

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, u3.String(), body)
		checkValueAndError(t, req, err)

		// execute login request.
		res, err = testClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		cookies := res.Cookies()
		require.Len(t, cookies, 1)
		assert.NotEqual(t, cookie, cookies[0])
	})

	test.Run("should be able to validate a 2FA token", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient(ctx, t)

		// create user.
		userInput := fakes.BuildFakeUserCreationInput()
		user, err := testClient.CreateUser(ctx, userInput)
		assert.NotNil(t, user)
		require.NoError(t, err)

		twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(user.TwoFactorQRCode)
		require.NoError(t, err)

		token, err := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)

		assert.NoError(t, testClient.VerifyTOTPSecret(ctx, user.ID, token))
	})

	test.Run("should reject attempt to validate an invalid 2FA token", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient(ctx, t)

		// create user.
		userInput := fakes.BuildFakeUserCreationInput()
		user, err := testClient.CreateUser(ctx, userInput)
		assert.NotNil(t, user)
		require.NoError(t, err)

		assert.Error(t, testClient.VerifyTOTPSecret(ctx, user.ID, "NOTREAL"))
	})

	test.Run("should be able to change 2FA Token", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)

		// create TOTP secret update request.
		token, err := totp.GenerateCode(testUser.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		ir := &types.TOTPSecretRefreshInput{
			CurrentPassword: testUser.HashedPassword,
			TOTPToken:       token,
		}
		out, err := json.Marshal(ir)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(testClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/totp_secret/new"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
		checkValueAndError(t, req, err)

		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})
		require.NoError(t, err)
		req.AddCookie(cookie)

		// execute TOTP secret update request.
		res, err := testClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		// load user response.
		r := &types.TOTPSecretRefreshResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(r))
		require.NotEqual(t, testUser.TwoFactorSecret, r.TwoFactorSecret)

		secretVerificationToken, err := totp.GenerateCode(r.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, secretVerificationToken, err)

		assert.NoError(t, testClient.VerifyTOTPSecret(ctx, testUser.ID, secretVerificationToken))

		// logout.
		assert.NoError(t, testClient.Logout(ctx))

		// create login request.
		newToken, err := totp.GenerateCode(r.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, newToken, err)

		secondCookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: newToken,
		})
		assert.NoError(t, err)
		assert.NotNil(t, secondCookie)
	})

	test.Run("should accept a cookie if a token is missing", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)
		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})
		require.NoError(t, err)

		// make arbitrary request.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testClient.BuildURL(nil, "webhooks"), nil)
		assert.NoError(t, err)
		req.AddCookie(cookie)

		res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	test.Run("should only allow clients with a given scope to see that scope's content", func(t *testing.T) {
		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)
		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})
		require.NoError(t, err)

		// create user.
		input := buildDummyOAuth2ClientInput(test, testUser.Username, testUser.HashedPassword, testUser.TwoFactorSecret)
		input.Scopes = []string{"absolutelynevergonnaexistascopelikethis"}
		premade, err := testClient.CreateOAuth2Client(ctx, cookie, input)
		checkValueAndError(test, premade, err)

		c, err := httpclient.NewClient(
			ctx,
			premade.ClientID,
			premade.ClientSecret,
			testClient.URL,
			noop.NewLogger(),
			buildHTTPClient(),
			premade.Scopes,
			true,
		)
		checkValueAndError(test, c, err)

		i, err := c.GetOAuth2Clients(ctx, nil)
		assert.Nil(t, i)
		assert.Error(t, err, "should experience error trying to fetch entry they're not authorized for")
	})
}
