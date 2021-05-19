package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(authTestSuite))
}

type authTestSuite struct {
	suite.Suite

	ctx           context.Context
	exampleUser   *types.User
	exampleCookie *http.Cookie
}

var _ suite.SetupTestSuite = (*authTestSuite)(nil)

func (s *authTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.exampleCookie = &http.Cookie{Name: s.T().Name()}

	s.exampleUser = fakes.BuildFakeUser()
	// the hashed passwords is never transmitted over the wire.
	s.exampleUser.HashedPassword = ""
	// the two factor secret is transmitted over the wire only on creation.
	s.exampleUser.TwoFactorSecret = ""
	// the two factor secret validation is never transmitted over the wire.
	s.exampleUser.TwoFactorSecretVerifiedOn = nil
}

func (s *authTestSuite) TestClient_UserStatus() {
	const expectedPath = "/auth/status"

	s.Run("standard", func() {
		t := s.T()

		expected := &types.UserStatusResponse{
			UserReputation:            s.exampleUser.ServiceAccountStatus,
			UserReputationExplanation: s.exampleUser.ReputationExplanation,
			UserIsAuthenticated:       true,
		}
		spec := newRequestSpec(false, http.MethodGet, "", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, expected)

		actual, err := c.UserStatus(s.ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.UserStatus(s.ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.UserStatus(s.ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func (s *authTestSuite) TestClient_Login() {
	const expectedPath = "/users/login"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserLoginInputFromUser(s.exampleUser)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				http.SetCookie(res, &http.Cookie{Name: s.exampleUser.Username})
			},
		))
		c := buildTestClient(t, ts)

		cookie, err := c.Login(s.ctx, exampleInput)
		require.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		cookie, err := c.Login(s.ctx, nil)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserLoginInputFromUser(s.exampleUser)

		c := buildTestClientWithInvalidURL(t)

		cookie, err := c.Login(s.ctx, exampleInput)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	s.Run("with timeout", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserLoginInputFromUser(s.exampleUser)
		c, _ := buildTestClientThatWaitsTooLong(t)

		cookie, err := c.Login(s.ctx, exampleInput)
		require.Nil(t, cookie)
		assert.Error(t, err)
	})

	s.Run("with missing cookie", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserLoginInputFromUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		cookie, err := c.Login(s.ctx, exampleInput)
		require.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func (s *authTestSuite) TestClient_Logout() {
	const expectedPath = "/users/logout"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		err := c.Logout(s.ctx)
		assert.NoError(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.Logout(s.ctx)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		err := c.Logout(s.ctx)
		assert.Error(t, err)
	})
}

func (s *authTestSuite) TestClient_ChangePassword() {
	const expectedPath = "/users/password/new"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		err := c.ChangePassword(s.ctx, s.exampleCookie, exampleInput)
		assert.NoError(t, err)
	})

	s.Run("with nil cookie", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		err := c.ChangePassword(s.ctx, nil, exampleInput)
		assert.Error(t, err)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.ChangePassword(s.ctx, s.exampleCookie, nil)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		err := c.ChangePassword(s.ctx, s.exampleCookie, exampleInput)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		err := c.ChangePassword(s.ctx, s.exampleCookie, exampleInput)
		assert.Error(t, err)
	})

	s.Run("with unsatisfactory response code", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusTeapot)
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		err := c.ChangePassword(s.ctx, s.exampleCookie, exampleInput)
		assert.Error(t, err)
	})
}

func (s *authTestSuite) TestClient_CycleTwoFactorSecret() {
	const expectedPath = "/users/totp_secret/new"

	s.Run("standard", func() {
		t := s.T()

		expected := &types.TOTPSecretRefreshResponse{
			TwoFactorQRCode: t.Name(),
			TwoFactorSecret: t.Name(),
		}
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, expected)
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := c.CycleTwoFactorSecret(s.ctx, s.exampleCookie, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	s.Run("with nil cookie", func() {
		t := s.T()

		expected := &types.TOTPSecretRefreshResponse{
			TwoFactorQRCode: t.Name(),
			TwoFactorSecret: t.Name(),
		}
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, expected)
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := c.CycleTwoFactorSecret(s.ctx, nil, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.CycleTwoFactorSecret(s.ctx, s.exampleCookie, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := &types.TOTPSecretRefreshInput{}

		actual, err := c.CycleTwoFactorSecret(s.ctx, s.exampleCookie, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := c.CycleTwoFactorSecret(s.ctx, s.exampleCookie, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := c.CycleTwoFactorSecret(s.ctx, s.exampleCookie, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func (s *authTestSuite) TestClient_VerifyTOTPSecret() {
	const expectedPath = "/users/totp_secret/verify"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)
	})

	s.Run("with invalid user ID", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildSimpleTestClient(t)

		err := c.VerifyTOTPSecret(s.ctx, 0, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	s.Run("with invalid token", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, " doesn't parse lol ")
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)

		c := buildTestClientWithInvalidURL(t)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	s.Run("with bad request response", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusBadRequest)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTOTPToken, err)
	})

	s.Run("with otherwise invalid status code response", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusInternalServerError)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	s.Run("with timeout", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		c.unauthenticatedClient.Timeout = time.Millisecond
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})
}
