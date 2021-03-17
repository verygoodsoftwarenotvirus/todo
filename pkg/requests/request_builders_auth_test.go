package requests

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestV1Client_BuildLoginRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/login"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		actual, err := c.BuildLoginRequest(ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		req, err := c.BuildLoginRequest(ctx, nil)
		assert.Nil(t, req)
		assert.Error(t, err)
	})
}

func TestV1Client_BuildVerifyTOTPSecretRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/verify"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		actual, err := c.BuildVerifyTOTPSecretRequest(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}
