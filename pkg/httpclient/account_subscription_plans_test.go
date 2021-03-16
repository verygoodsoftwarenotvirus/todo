package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	exampleInput                       *types.AccountSubscriptionPlanCreationInput
	exampleAccountSubscriptionPlanList *types.AccountSubscriptionPlanList
}

var _ suite.SetupTestSuite = (*accountSubscriptionPlansTestSuite)(nil)

func (s *accountSubscriptionPlansTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	s.exampleInput = fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)
	s.exampleAccountSubscriptionPlanList = fakes.BuildFakeAccountSubscriptionPlanList()
}

func (s *accountSubscriptionPlansTestSuite) TestV1Client_GetAccountSubscriptionPlan() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleAccountSubscriptionPlan))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAccountSubscriptionPlan, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountSubscriptionPlansTestSuite) TestV1Client_GetAccountSubscriptionPlans() {
	const expectedPath = "/api/v1/account_subscription_plans"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("happy path", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleAccountSubscriptionPlanList))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlans(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAccountSubscriptionPlanList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccountSubscriptionPlans(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlans(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountSubscriptionPlansTestSuite) TestV1Client_CreateAccountSubscriptionPlan() {
	const expectedPath = "/api/v1/account_subscription_plans"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				var x *types.AccountSubscriptionPlanCreationInput
				require.NoError(t, json.NewDecoder(req.Body).Decode(&x))

				assert.Equal(t, s.exampleInput, x)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleAccountSubscriptionPlan))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, s.exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAccountSubscriptionPlan, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateAccountSubscriptionPlan(s.ctx, s.exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountSubscriptionPlansTestSuite) TestV1Client_UpdateAccountSubscriptionPlan() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				assert.NoError(t, json.NewEncoder(res).Encode(s.exampleAccountSubscriptionPlan))
			},
		))

		err := buildTestClient(t, ts).UpdateAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).UpdateAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountSubscriptionPlansTestSuite) TestV1Client_ArchiveAccountSubscriptionPlan() {
	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccountSubscriptionPlan.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusOK)
			},
		))

		err := buildTestClient(t, ts).ArchiveAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *accountSubscriptionPlansTestSuite) TestV1Client_GetAuditLogForAccountSubscriptionPlan() {
	const (
		expectedPath   = "/api/v1/account_subscription_plans/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccountSubscriptionPlan.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccountSubscriptionPlan.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(s.ctx, s.exampleAccountSubscriptionPlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
