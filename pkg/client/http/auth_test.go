package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(authTestSuite))
}

type authTestSuite struct {
	suite.Suite

	ctx         context.Context
	exampleUser *types.User
}

var _ suite.SetupTestSuite = (*authTestSuite)(nil)

func (s *authTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.exampleUser = fakes.BuildFakeUser()
	// the hashed authentication is never transmitted over the wire.
	s.exampleUser.HashedPassword = ""
	// the two factor secret is transmitted over the wire only on creation.
	s.exampleUser.TwoFactorSecret = ""
	// the two factor secret validation is never transmitted over the wire.
	s.exampleUser.TwoFactorSecretVerifiedOn = nil
}

func (s *authTestSuite) TestV1Client_Login() {
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

	s.Run("with invalid client url", func() {
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
		c, _ := buildTestClientThatWaitsTooLong(t, spec)

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

func (s *authTestSuite) TestV1Client_VerifyTOTPSecret() {
	const expectedPath = "/users/totp_secret/verify"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)
	})

	s.Run("with bad request response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusBadRequest)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTOTPToken, err)
	})

	s.Run("with otherwise invalid status code response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusInternalServerError)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)

		c := buildTestClientWithInvalidURL(t)

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	s.Run("with timeout", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)

		c, _ := buildTestClientThatWaitsTooLong(t, spec)
		c.unauthenticatedClient.Timeout = time.Millisecond

		err := c.VerifyTOTPSecret(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})
}
