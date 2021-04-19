package requests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(authTestSuite))
}

type authTestSuite struct {
	suite.Suite

	ctx         context.Context
	builder     *Builder
	exampleUser *types.User
}

var _ suite.SetupTestSuite = (*authTestSuite)(nil)

func (s *authTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()

	s.exampleUser = fakes.BuildFakeUser()
	// the hashed passwords is never transmitted over the wire.
	s.exampleUser.HashedPassword = ""
	// the two factor secret is transmitted over the wire only on creation.
	s.exampleUser.TwoFactorSecret = ""
	// the two factor secret validation is never transmitted over the wire.
	s.exampleUser.TwoFactorSecretVerifiedOn = nil
}

func (s *authTestSuite) TestBuilder_BuildLoginRequest() {
	const expectedPath = "/users/login"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(s.exampleUser)

		actual, err := s.builder.BuildLoginRequest(s.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	s.Run("with nil input", func() {
		t := s.T()

		req, err := s.builder.BuildLoginRequest(s.ctx, nil)
		assert.Nil(t, req)
		assert.Error(t, err)
	})
}

func (s *authTestSuite) TestBuilder_BuildVerifyTOTPSecretRequest() {
	const expectedPath = "/users/totp_secret/verify"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(s.exampleUser)

		actual, err := s.builder.BuildVerifyTOTPSecretRequest(s.ctx, s.exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}
