package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
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

func (s *adminTestSuite) TestV1Client_UpdateUserReputation() {
	const expectedPath = "/api/v1/_admin_/users/status"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		err := c.UpdateUserReputation(s.ctx, exampleInput)
		assert.NoError(t, err)
	})

	s.Run("with bad request response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusBadRequest)

		assert.Error(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})

	s.Run("with otherwise invalid status code response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusInternalServerError)

		err := c.UpdateUserReputation(s.ctx, exampleInput)
		assert.Error(t, err)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		c := buildTestClientWithInvalidURL(t)
		err := c.UpdateUserReputation(s.ctx, exampleInput)

		assert.Error(t, err)
	})

	s.Run("with timeout", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientThatWaitsTooLong(t, spec)

		err := c.UpdateUserReputation(s.ctx, exampleInput)
		assert.Error(t, err)
	})
}
