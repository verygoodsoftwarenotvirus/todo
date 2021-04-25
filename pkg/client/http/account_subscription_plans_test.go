package http

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

func TestAccountSubscriptionPlans(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(accountSubscriptionPlansTestSuite))
}

type accountSubscriptionPlansTestSuite struct {
	suite.Suite

	ctx                                context.Context
	exampleAccountSubscriptionPlan     *types.AccountSubscriptionPlan
	exampleAccountSubscriptionPlanList *types.AccountSubscriptionPlanList
}

var _ suite.SetupTestSuite = (*accountSubscriptionPlansTestSuite)(nil)

func (s *accountSubscriptionPlansTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	s.exampleAccountSubscriptionPlanList = fakes.BuildFakeAccountSubscriptionPlanList()
}

func (s *accountSubscriptionPlansTestSuite) TestClient_GetAccountSubscriptionPlan() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAccountSubscriptionPlan)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.NoError(t, err)
		assert.Equal(t, s.exampleAccountSubscriptionPlan, actual)
	})

	s.Run("returns error with zero ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, 0)

		assert.Error(t, err)
		assertErrorMatches(t, err, ErrInvalidIDProvided)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)
		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *accountSubscriptionPlansTestSuite) TestClient_GetAccountSubscriptionPlans() {
	const expectedPath = "/api/v1/account_subscription_plans"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		filter := types.DefaultQueryFilter()

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAccountSubscriptionPlanList)
		actual, err := c.GetAccountSubscriptionPlans(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAccountSubscriptionPlanList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccountSubscriptionPlans(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAccountSubscriptionPlans(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *accountSubscriptionPlansTestSuite) TestClient_CreateAccountSubscriptionPlan() {
	const expectedPath = "/api/v1/account_subscription_plans"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c := buildTestClientWithRequestBodyValidation(t, spec, &types.AccountSubscriptionPlanCreationInput{}, exampleInput, s.exampleAccountSubscriptionPlan)
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, exampleInput)

		assert.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, s.exampleAccountSubscriptionPlan, actual)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, nil)

		require.Nil(t, actual)
		assert.Error(t, err)
		assertErrorMatches(t, err, ErrNilInputProvided)
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := &types.AccountSubscriptionPlanCreationInput{}
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, exampleInput)

		require.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)
		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with request failure", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)
		c, _ := buildTestClientThatWaitsTooLong(t)
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, exampleInput)

		require.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *accountSubscriptionPlansTestSuite) TestClient_UpdateAccountSubscriptionPlan() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAccountSubscriptionPlan)

		err := c.UpdateAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan)
		assert.NoError(t, err)
	})

	s.Run("returns error with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		err := c.UpdateAccountSubscriptionPlan(s.ctx, nil)

		assert.Error(t, err)
		assertErrorMatches(t, err, ErrNilInputProvided)
	})

	s.Run("with error building request", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).UpdateAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan)
		assert.Error(t, err)
	})

	s.Run("with request failure", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		err := c.UpdateAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan)

		assert.Error(t, err)
	})
}

func (s *accountSubscriptionPlansTestSuite) TestClient_ArchiveAccountSubscriptionPlan() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)
	})

	s.Run("returns error with zero ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.ArchiveAccountSubscriptionPlan(s.ctx, 0)
		assert.Error(t, err)
		assertErrorMatches(t, err, ErrInvalidIDProvided)
	})

	s.Run("with error building request", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		err := c.ArchiveAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err)
	})
}

func (s *accountSubscriptionPlansTestSuite) TestClient_GetAuditLogForAccountSubscriptionPlan() {
	const (
		expectedPath   = "/api/v1/account_subscription_plans/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccountSubscriptionPlan.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		c, _ := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with zero ID input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, 0)

		assert.Nil(t, actual)
		assert.Error(t, err)
		assertErrorMatches(t, err, ErrInvalidIDProvided)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccountSubscriptionPlan.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
