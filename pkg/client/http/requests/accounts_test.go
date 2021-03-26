package requests

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	builder            *Builder
	exampleAccount     *types.Account
	exampleInput       *types.AccountCreationInput
	exampleAccountList *types.AccountList
}

var _ suite.SetupTestSuite = (*accountRequestBuildersTestSuite)(nil)

func (s *accountRequestBuildersTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleInput = fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)
	s.exampleAccountList = fakes.BuildFakeAccountList()
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildGetAccountRequest() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccount.ID)

		actual, err := s.builder.BuildGetAccountRequest(s.ctx, s.exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildGetAccountsRequest() {
	const expectedPath = "/api/v1/accounts"

	s.Run("standard", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := s.builder.BuildGetAccountsRequest(s.ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildCreateAccountRequest() {
	const expectedPath = "/api/v1/accounts"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildCreateAccountRequest(s.ctx, s.exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildUpdateAccountRequest() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccount.ID)

		actual, err := s.builder.BuildUpdateAccountRequest(s.ctx, s.exampleAccount)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildArchiveAccountRequest() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccount.ID)

		actual, err := s.builder.BuildArchiveAccountRequest(s.ctx, s.exampleAccount.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildRemoveUserRequest() {
	const expectedPathFormat = "/api/v1/accounts/%d/members/%d"

	s.Run("standard", func() {
		t := s.T()

		reason := t.Name()
		expectedReason := url.QueryEscape(reason)
		spec := newRequestSpec(true, http.MethodDelete, fmt.Sprintf("reason=%s", expectedReason), expectedPathFormat, s.exampleAccount.ID, s.exampleAccount.BelongsToUser)

		actual, err := s.builder.BuildRemoveUserRequest(s.ctx, s.exampleAccount.ID, s.exampleAccount.BelongsToUser, reason)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountRequestBuildersTestSuite) TestBuilder_BuildGetAuditLogForAccountRequest() {
	const expectedPath = "/api/v1/accounts/%d/audit"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildGetAuditLogForAccountRequest(s.ctx, s.exampleAccount.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleAccount.ID)
		assertRequestQuality(t, actual, spec)
	})
}
