package requests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func TestAccountSubscriptionPlans(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(accountSubscriptionPlansRequestBuildersTestSuite))
}

type accountSubscriptionPlansRequestBuildersTestSuite struct {
	suite.Suite

	ctx                                context.Context
	builder                            *Builder
	exampleAccountSubscriptionPlan     *types.AccountSubscriptionPlan
	exampleInput                       *types.AccountSubscriptionPlanCreationInput
	exampleAccountSubscriptionPlanList *types.AccountSubscriptionPlanList
}

var _ suite.SetupTestSuite = (*accountSubscriptionPlansRequestBuildersTestSuite)(nil)

func (s *accountSubscriptionPlansRequestBuildersTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()
	s.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	s.exampleInput = fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)
	s.exampleAccountSubscriptionPlanList = fakes.BuildFakeAccountSubscriptionPlanList()
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestBuilder_BuildGetAccountSubscriptionPlanRequest() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		actual, err := s.builder.BuildGetAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestBuilder_BuildGetAccountSubscriptionPlansRequest() {
	const expectedPath = "/api/v1/account_subscription_plans"

	s.Run("standard", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := s.builder.BuildGetAccountSubscriptionPlansRequest(s.ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestBuilder_BuildCreateAccountSubscriptionPlanRequest() {
	const expectedPath = "/api/v1/account_subscription_plans"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildCreateAccountSubscriptionPlanRequest(s.ctx, s.exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestBuilder_BuildUpdateAccountSubscriptionPlanRequest() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		actual, err := s.builder.BuildUpdateAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestBuilder_BuildArchiveAccountSubscriptionPlanRequest() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		actual, err := s.builder.BuildArchiveAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestBuilder_BuildGetAuditLogForAccountSubscriptionPlanRequest() {
	const expectedPath = "/api/v1/account_subscription_plans/%d/audit"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildGetAuditLogForAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleAccountSubscriptionPlan.ID)
		assertRequestQuality(t, actual, spec)
	})
}
