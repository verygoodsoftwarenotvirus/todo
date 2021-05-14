package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestLogin() {
	s.Run("logging in and out works", func() {
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

		assert.Equal(t, auth.DefaultCookieName, cookie.Name)
		assert.NotEmpty(t, cookie.Value)
		assert.NotZero(t, cookie.MaxAge)
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)

		assert.NoError(t, testClient.Logout(ctx))
	})
}

func (s *TestSuite) TestLogin_WithoutBodyFails() {
	s.Run("login request without body fails", func() {
		t := s.T()

		ctx, span := tracing.StartSpan(context.Background())
		defer span.End()

		_, _, testClient, _ := createUserAndClientForTest(ctx, t)

		u, err := url.Parse(testClient.BuildURL(ctx, nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
		requireNotNilAndNoProblems(t, req, err)

		// execute login request.
		res, err := testClient.PlainClient().Do(req)
		requireNotNilAndNoProblems(t, res, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func (s *TestSuite) TestLogin_ShouldNotBeAbleToLoginWithInvalidPassword() {
	s.Run("should not be able to log in with the wrong password", func() {
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
}

func (s *TestSuite) TestLogin_ShouldNotBeAbleToLoginAsANonexistentUser() {
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
}

func (s *TestSuite) TestLogin_ShouldNotBeAbleToLoginWithoutValidating2FASecret() {
	s.Run("should not be able to login without validating 2FA secret", func() {
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

		actual, err := testClient.UserStatus(ctx)
		assert.NoError(t, err)

		assert.Equal(t, true, actual.UserIsAuthenticated, "expected UserIsAuthenticated to equal %v, but got %v", true, actual.UserIsAuthenticated)
		assert.Equal(t, types.GoodStandingAccountStatus, actual.UserReputation, "expected UserReputation to equal %v, but got %v", types.GoodStandingAccountStatus, actual.UserReputation)
		assert.Equal(t, "", actual.UserReputationExplanation, "expected UserReputationExplanation to equal %v, but got %v", "", actual.UserReputationExplanation)
		assert.Equal(t, &permissions.ServiceAdminPermissionsSummary{}, actual.ServiceAdminPermissionsSummary, "expected ServiceAdminPermissionsSummary to be nil, but got %v", actual.ServiceAdminPermissionsSummary)
		assert.NotZero(t, actual.ActiveAccount)

		var activeAccountPresent bool
		for _, accountInfo := range actual.AccountPermissions {
			if !activeAccountPresent && accountInfo.AccountID == actual.ActiveAccount {
				activeAccountPresent = true
				break
			}
		}
		assert.True(t, activeAccountPresent)

		assert.NotEmpty(t, actual.AccountPermissions)

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

		req, err := testClient.RequestBuilder().BuildAPIClientAuthTokenRequest(ctx, input, actualKey)
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

		// create new passwords.
		var backwardsPass string
		for _, v := range testUser.HashedPassword {
			backwardsPass = string(v) + backwardsPass
		}

		// update passwords.
		assert.NoError(t, testClient.ChangePassword(ctx, cookie, &types.PasswordUpdateInput{
			CurrentPassword: testUser.HashedPassword,
			TOTPToken:       generateTOTPTokenForUser(t, testUser),
			NewPassword:     backwardsPass,
		}))

		// logout.
		assert.NoError(t, testClient.Logout(ctx))

		// login again with new passwords.
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

		assert.NoError(t, testClient.VerifyTOTPSecret(ctx, user.CreatedUserID, token))
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

		assert.Error(t, testClient.VerifyTOTPSecret(ctx, user.CreatedUserID, "NOTREAL"))
	})
}
