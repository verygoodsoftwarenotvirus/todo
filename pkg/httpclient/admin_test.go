package httpclient

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

func TestAdmin(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(adminTestSuite))
}

type adminTestSuite struct {
	suite.Suite

	ctx                context.Context
	exampleAccount     *types.Account
	exampleInput       *types.AccountCreationInput
	exampleAccountList *types.AccountList
}

var _ suite.SetupTestSuite = (*adminTestSuite)(nil)

func (s *adminTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleInput = fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)
	s.exampleAccountList = fakes.BuildFakeAccountList()
}

func (s *adminTestSuite) TestV1Client_BuildBanUserRequest() {
	const expectedPathFormat = "/api/v1/_admin_/users/status"

	s.Run("happy path", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildAccountStatusUpdateInputRequest(s.ctx, exampleInput)
		assert.NoError(t, err)
		require.NotNil(t, actual)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *adminTestSuite) TestV1Client_BanUser() {
	const expectedPathFormat = "/api/v1/_admin_/users/status"

	s.Run("happy path", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusAccepted)
			},
		))
		c := buildTestClient(t, ts)

		err := c.UpdateAccountStatus(s.ctx, exampleInput)
		assert.NoError(t, err)
	})

	s.Run("with bad request response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusBadRequest)
			},
		))
		c := buildTestClient(t, ts)

		assert.Error(t, c.UpdateAccountStatus(s.ctx, exampleInput))
	})

	s.Run("with otherwise invalid status code response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusInternalServerError)
			},
		))
		c := buildTestClient(t, ts)

		err := c.UpdateAccountStatus(s.ctx, exampleInput)
		assert.Error(t, err)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		c := buildTestClientWithInvalidURL(t)
		err := c.UpdateAccountStatus(s.ctx, exampleInput)

		assert.Error(t, err)
	})

	s.Run("with timeout", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				time.Sleep(10 * time.Minute)

				res.WriteHeader(http.StatusAccepted)
			},
		))
		c := buildTestClient(t, ts)
		require.NoError(t, c.SetOptions(UsingTimeout(time.Millisecond)))

		err := c.UpdateAccountStatus(s.ctx, exampleInput)
		assert.Error(t, err)
	})
}
