package requests

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	exampleAccountSubscriptionPlan     *types.AccountSubscriptionPlan
	exampleInput                       *types.AccountSubscriptionPlanCreationInput
	exampleAccountSubscriptionPlanList *types.AccountSubscriptionPlanList
}

var _ suite.SetupTestSuite = (*accountSubscriptionPlansRequestBuildersTestSuite)(nil)

func (s *accountSubscriptionPlansRequestBuildersTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	s.exampleInput = fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)
	s.exampleAccountSubscriptionPlanList = fakes.BuildFakeAccountSubscriptionPlanList()
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestV1Client_BuildGetAccountSubscriptionPlanRequest() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestV1Client_BuildGetAccountSubscriptionPlansRequest() {
	const expectedPath = "/api/v1/account_subscription_plans"

	s.Run("happy path", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetAccountSubscriptionPlansRequest(s.ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestV1Client_BuildCreateAccountSubscriptionPlanRequest() {
	const expectedPath = "/api/v1/account_subscription_plans"

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateAccountSubscriptionPlanRequest(s.ctx, s.exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestV1Client_BuildUpdateAccountSubscriptionPlanRequest() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestV1Client_BuildArchiveAccountSubscriptionPlanRequest() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *accountSubscriptionPlansRequestBuildersTestSuite) TestV1Client_BuildGetAuditLogForAccountSubscriptionPlanRequest() {
	const expectedPath = "/api/v1/account_subscription_plans/%d/audit"

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForAccountSubscriptionPlanRequest(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleAccountSubscriptionPlan.ID)
		assertRequestQuality(t, actual, spec)
	})
}
