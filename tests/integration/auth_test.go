package integration

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestAuth(test *testing.T) {
	test.Parallel()

	test.Run("should be able to login and log out", func(t *testing.T) {
		t.Parallel()

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

		assert.Equal(t, authservice.DefaultCookieName, cookie.Name)
		assert.NotEmpty(t, cookie.Value)
		assert.NotZero(t, cookie.MaxAge)
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)

		assert.NoError(t, testClient.Logout(ctx))
	})

	test.Run("should be able to check their auth status with the client", func(t *testing.T) {
		t.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)
		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})

		require.NotNil(t, cookie)
		assert.NoError(t, err)

		actual, err := testClient.Status(ctx, cookie)
		assert.NoError(t, err)

		expected := &types.UserStatusResponse{
			UserIsAuthenticated:      true,
			UserIsAdmin:              false,
			UserAccountStatus:        types.GoodStandingAccountStatus,
			AccountStatusExplanation: "",
			AdminPermissions:         nil,
		}

		assert.Equal(t, expected, actual)
		assert.NoError(t, err)

		assert.NoError(t, testClient.Logout(ctx))
	})

	test.Run("login request without body fails", func(t *testing.T) {
		t.Parallel()

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
		t.Parallel()

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
		t.Parallel()

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
		t.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient()

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
		t.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testClient.BuildURL(nil, "webhooks"), nil)
		assert.NoError(t, err)

		res, err := testClient.PlainClient().Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	test.Run("should be able to change password", func(t *testing.T) {
		t.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, testClient := createUserAndClientForTest(ctx, t)

		// login.
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

		// update password.
		assert.NoError(t, testClient.ChangePassword(ctx, cookie, &types.PasswordUpdateInput{
			CurrentPassword: testUser.HashedPassword,
			TOTPToken:       generateTOTPTokenForUser(t, testUser),
			NewPassword:     backwardsPass,
		}))

		// logout.
		assert.NoError(t, testClient.Logout(ctx))

		// login again with new password.
		cookie, err = testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  backwardsPass,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})

		assert.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	test.Run("should be able to validate a 2FA token", func(t *testing.T) {
		t.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient()

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
		t.Parallel()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient()

		// create user.
		userInput := fakes.BuildFakeUserCreationInput()
		user, err := testClient.CreateUser(ctx, userInput)
		assert.NotNil(t, user)
		require.NoError(t, err)

		assert.Error(t, testClient.VerifyTOTPSecret(ctx, user.ID, "NOTREAL"))
	})

	test.Run("should be able to change 2FA Token", func(t *testing.T) {
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

		r, err := testClient.CycleTwoFactorSecret(ctx, cookie, &types.TOTPSecretRefreshInput{
			CurrentPassword: testUser.HashedPassword,
			TOTPToken:       generateTOTPTokenForUser(t, testUser),
		})
		assert.NoError(t, err)

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

		// make arbitrary request.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testClient.BuildURL(nil, "webhooks"), nil)
		assert.NoError(t, err)
		req.AddCookie(cookie)

		res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	test.Run("should only allow clients with a given scope to see that scope's content", func(t *testing.T) {
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

		// create user.
		input := buildDummyOAuth2ClientInput(test, testUser.Username, testUser.HashedPassword, testUser.TwoFactorSecret)
		input.Scopes = []string{"absolutelynevergonnaexistascopelikethis"}
		premade, err := testClient.CreateOAuth2Client(ctx, cookie, input)
		checkValueAndError(test, premade, err)

		c := httpclient.NewClient(
			httpclient.WithURL(testClient.URL()),
			httpclient.WithLogger(noop.NewLogger()),
			httpclient.WithOAuth2ClientCredentials(
				httpclient.BuildClientCredentialsConfig(
					testClient.URL(),
					premade.ClientID,
					premade.ClientSecret,
					premade.Scopes...,
				),
			),
			httpclient.WithDebugEnabled(),
		)
		checkValueAndError(test, c, nil)

		i, err := c.GetOAuth2Clients(ctx, nil)
		assert.Nil(t, i)
		assert.Error(t, err, "should experience error trying to fetch entry they're not authorized for")
	})
}
