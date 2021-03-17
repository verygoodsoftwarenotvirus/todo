package httpclient

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccounts(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(accountsTestSuite))
}

type accountsTestSuite struct {
	suite.Suite

	ctx                context.Context
	exampleAccount     *types.Account
	exampleInput       *types.AccountCreationInput
	exampleAccountList *types.AccountList
}

var _ suite.SetupTestSuite = (*accountsTestSuite)(nil)

func (s *accountsTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleInput = fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)
	s.exampleAccountList = fakes.BuildFakeAccountList()
}

func (s *accountsTestSuite) TestV1Client_GetAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccount.ID)

		c := buildTestClientWithJSONResponse(t, spec, s.exampleAccount)
		actual, err := c.GetAccount(s.ctx, s.exampleAccount.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAccount, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccount(s.ctx, s.exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccount.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAccount(s.ctx, s.exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountsTestSuite) TestV1Client_GetAccounts() {
	const expectedPath = "/api/v1/accounts"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
	filter := (*types.QueryFilter)(nil)

	s.Run("happy path", func() {
		t := s.T()

		c := buildTestClientWithJSONResponse(t, spec, s.exampleAccountList)
		actual, err := c.GetAccounts(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAccountList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccounts(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAccounts(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountsTestSuite) TestV1Client_CreateAccount() {
	const expectedPath = "/api/v1/accounts"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("happy path", func() {
		t := s.T()

		s.exampleAccount.BelongsToUser = 0
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)

		c := buildTestClientWithRequestBodyValidation(t, spec, &types.AccountCreationInput{}, exampleInput, s.exampleAccount)
		actual, err := c.CreateAccount(s.ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAccount, actual)
	})

	s.Run("with invalid client URL", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateAccount(s.ctx, s.exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountsTestSuite) TestV1Client_UpdateAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccount.ID)

		err := buildTestClientWithJSONResponse(t, spec, s.exampleAccount).UpdateAccount(s.ctx, s.exampleAccount)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client URL", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.UpdateAccount(s.ctx, s.exampleAccount)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountsTestSuite) TestV1Client_ArchiveAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccount.ID)
		c := buildTestClientWithOKResponse(t, spec)

		err := c.ArchiveAccount(s.ctx, s.exampleAccount.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveAccount(s.ctx, s.exampleAccount.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountsTestSuite) TestV1Client_GetAuditLogForAccount() {
	const (
		expectedPath   = "/api/v1/accounts/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("happy path", func() {
		t := s.T()

		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccount.ID)

		c := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)
		actual, err := c.GetAuditLogForAccount(s.ctx, s.exampleAccount.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAccount(s.ctx, s.exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccount.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAuditLogForAccount(s.ctx, s.exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
