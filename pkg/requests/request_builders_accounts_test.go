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
	"github.com/stretchr/testify/require"
)

func TestAccountRequestBuilders(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(accountRequestBuildersTestSuite))
}

type accountRequestBuildersTestSuite struct {
	suite.Suite

	ctx                context.Context
	exampleAccount     *types.Account
	exampleInput       *types.AccountCreationInput
	exampleAccountList *types.AccountList
}

var _ suite.SetupTestSuite = (*accountRequestBuildersTestSuite)(nil)

func (s *accountRequestBuildersTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleInput = fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)
	s.exampleAccountList = fakes.BuildFakeAccountList()
}

func (s *accountRequestBuildersTestSuite) TestV1Client_BuildGetAccountRequest() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccount.ID)

		c := buildTestClientWithNilServer(t)
		actual, err := c.BuildGetAccountRequest(s.ctx, s.exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAccountsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetAccountsRequest(ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildCreateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateAccountRequest(ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildUpdateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleAccount.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateAccountRequest(ctx, exampleAccount)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildArchiveAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		exampleAccount := fakes.BuildFakeAccount()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccount.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveAccountRequest(ctx, exampleAccount.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAuditLogForAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForAccountRequest(ctx, exampleAccount.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAccount.ID)
		assertRequestQuality(t, actual, spec)
	})
}
