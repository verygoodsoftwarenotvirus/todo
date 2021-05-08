package httpclient

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
	exampleAccountList *types.AccountList
}

var _ suite.SetupTestSuite = (*adminTestSuite)(nil)

func (s *adminTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccountList = fakes.BuildFakeAccountList()
}

func (s *adminTestSuite) TestClient_UpdateUserReputation() {
	const expectedPath = "/api/v1/admin/users/status"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserReputationUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		assert.NoError(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})

	s.Run("with nil input", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		assert.Error(t, c.UpdateUserReputation(s.ctx, nil))
	})

	s.Run("with invalid input", func() {
		t := s.T()

		exampleInput := &types.UserReputationUpdateInput{}
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)

		assert.Error(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})

	s.Run("with bad request response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserReputationUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusBadRequest)

		assert.Error(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})

	s.Run("with otherwise invalid status code response", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserReputationUpdateInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusInternalServerError)

		assert.Error(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserReputationUpdateInput()
		c := buildTestClientWithInvalidURL(t)

		assert.Error(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})

	s.Run("with timeout", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserReputationUpdateInput()
		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.UpdateUserReputation(s.ctx, exampleInput))
	})
}
