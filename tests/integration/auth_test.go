package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

func (s *TestSuite) TestLogin() {
	s.Run("logging in and out works via cookie", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		testUser, _, testClient, _ := createUserAndClientForTest(ctx, t)
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

	s.Run("login request without body fails", func() {
		t := s.T()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		_, _, testClient, _ := createUserAndClientForTest(ctx, t)

		u, err := url.Parse(testClient.BuildURL(nil, nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
		requireNotNilAndNoProblems(t, req, err)

		// execute login request.
		res, err := testClient.PlainClient().Do(req)
		requireNotNilAndNoProblems(t, res, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	s.Run("should not be able to log in with the wrong authentication", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		testUser, _, testClient, _ := createUserAndClientForTest(ctx, t)

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

	s.Run("should not be able to login as someone that does not exist", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		testUser, _, testClient, _ := createUserAndClientForTest(ctx, t)

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

	s.Run("should not be able to login without validating TOTP secret", func() {
		t := s.T()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient(t)

		// create a user.
		exampleUser := fakes.BuildFakeUser()
		exampleUserCreationInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
		ucr, err := testClient.CreateUser(ctx, exampleUserCreationInput)
		requireNotNilAndNoProblems(t, ucr, err)

		twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(ucr.TwoFactorQRCode)
		require.NoError(t, err)

		// create login request.
		token, err := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
		requireNotNilAndNoProblems(t, token, err)
		r := &types.UserLoginInput{
			Username:  exampleUserCreationInput.Username,
			Password:  exampleUserCreationInput.Password,
			TOTPToken: token,
		}

		cookie, err := testClient.Login(ctx, r)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func (s *TestSuite) TestCheckingAuthStatus() {
	s.Run("checking auth status", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		testUser, _, testClient, _ := createUserAndClientForTest(ctx, t)
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
			UserAccountStatus:        types.GoodStandingAccountStatus,
			AccountStatusExplanation: "",
			ServiceAdminPermissions:  nil,
		}

		assert.Equal(t, expected, actual)
		assert.NoError(t, err)

		assert.NoError(t, testClient.Logout(ctx))
	})
}

func (s *TestSuite) TestPASETOGeneration() {
	s.Run("checking auth status", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		user, cookie, testClient, _ := createUserAndClientForTest(ctx, t)

		// Create API client.
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
		exampleAPIClientInput.UserLoginInput = types.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, user),
		}

		createdAPIClient, apiClientCreationErr := testClient.CreateAPIClient(ctx, cookie, exampleAPIClientInput)
		requireNotNilAndNoProblems(t, createdAPIClient, apiClientCreationErr)

		actualKey, keyDecodeErr := base64.RawURLEncoding.DecodeString(createdAPIClient.ClientSecret)
		require.NoError(t, keyDecodeErr)

		input := &types.PASETOCreationInput{
			ClientID:    createdAPIClient.ClientID,
			RequestTime: time.Now().UTC().UnixNano(),
		}

		req, err := testClient.BuildAPIClientAuthTokenRequest(ctx, input, actualKey)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, res)

		var tokenRes types.PASETOResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&tokenRes))

		assert.NotEmpty(t, tokenRes.Token)
		assert.NotEmpty(t, tokenRes.ExpiresAt)
	})
}

func (s *TestSuite) TestPasswordChanging() {
	s.Run("should be possible to change your password", func() {
		t := s.T()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testUser, _, testClient, _ := createUserAndClientForTest(ctx, t)

		// login.
		cookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})
		require.NotNil(t, cookie)
		assert.NoError(t, err)

		// create new authentication.
		var backwardsPass string
		for _, v := range testUser.HashedPassword {
			backwardsPass = string(v) + backwardsPass
		}

		// update authentication.
		assert.NoError(t, testClient.ChangePassword(ctx, cookie, &types.PasswordUpdateInput{
			CurrentPassword: testUser.HashedPassword,
			TOTPToken:       generateTOTPTokenForUser(t, testUser),
			NewPassword:     backwardsPass,
		}))

		// logout.
		assert.NoError(t, testClient.Logout(ctx))

		// login again with new authentication.
		cookie, err = testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  backwardsPass,
			TOTPToken: generateTOTPTokenForUser(t, testUser),
		})

		assert.NotNil(t, cookie)
		assert.NoError(t, err)
	})
}

func (s *TestSuite) TestTOTPSecretChanging() {
	s.Run("should be possible to change your TOTP secret", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		testUser, _, testClient, _ := createUserAndClientForTest(ctx, t)

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
		requireNotNilAndNoProblems(t, secretVerificationToken, err)

		assert.NoError(t, testClient.VerifyTOTPSecret(ctx, testUser.ID, secretVerificationToken))

		// logout.
		assert.NoError(t, testClient.Logout(ctx))

		// create login request.
		newToken, err := totp.GenerateCode(r.TwoFactorSecret, time.Now().UTC())
		requireNotNilAndNoProblems(t, newToken, err)

		secondCookie, err := testClient.Login(ctx, &types.UserLoginInput{
			Username:  testUser.Username,
			Password:  testUser.HashedPassword,
			TOTPToken: newToken,
		})
		assert.NoError(t, err)
		assert.NotNil(t, secondCookie)
	})
}

func (s *TestSuite) TestTOTPTokenValidation() {
	s.Run("should be possible to validate TOTP", func() {
		t := s.T()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		testClient := buildSimpleClient(t)

		// create user.
		userInput := fakes.BuildFakeUserCreationInput()
		user, err := testClient.CreateUser(ctx, userInput)
		assert.NotNil(t, user)
		require.NoError(t, err)

		twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(user.TwoFactorQRCode)
		require.NoError(t, err)

		token, err := totp.GenerateCode(twoFactorSecret, time.Now().UTC())
		requireNotNilAndNoProblems(t, token, err)

		assert.NoError(t, testClient.VerifyTOTPSecret(ctx, user.ID, token))
	})

	s.Run("should not be possible to validate an invalid TOTP", func() {
		t := s.T()

		ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
		defer span.End()

		testClient := buildSimpleClient(t)

		// create user.
		userInput := fakes.BuildFakeUserCreationInput()
		user, err := testClient.CreateUser(ctx, userInput)
		assert.NotNil(t, user)
		require.NoError(t, err)

		assert.Error(t, testClient.VerifyTOTPSecret(ctx, user.ID, "NOTREAL"))
	})
}
