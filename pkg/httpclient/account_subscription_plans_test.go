package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, examplePlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetAccountSubscriptionPlanRequest(ctx, examplePlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, examplePlan.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(examplePlan))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlan(ctx, examplePlan.ID)

		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, examplePlan, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccountSubscriptionPlan(ctx, examplePlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, examplePlan.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlan(ctx, examplePlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetAccountSubscriptionPlansRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetAccountSubscriptionPlansRequest(ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAccountSubscriptionPlans(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		examplePlanList := fakes.BuildFakePlanList()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(examplePlanList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, examplePlanList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildCreateAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(examplePlan)

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateAccountSubscriptionPlanRequest(ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_CreateAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(examplePlan)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					var x *types.AccountSubscriptionPlanCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))

					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(examplePlan))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, examplePlan, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(examplePlan)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildUpdateAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, examplePlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateAccountSubscriptionPlanRequest(ctx, examplePlan)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_UpdateAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, examplePlan.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					assert.NoError(t, json.NewEncoder(res).Encode(examplePlan))
				},
			),
		)

		err := buildTestClient(t, ts).UpdateAccountSubscriptionPlan(ctx, examplePlan)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		err := buildTestClientWithInvalidURL(t).UpdateAccountSubscriptionPlan(ctx, examplePlan)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildArchiveAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, examplePlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveAccountSubscriptionPlanRequest(ctx, examplePlan.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_ArchiveAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, examplePlan.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		err := buildTestClientWithInvalidURL(t).ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetAuditLogForAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx, examplePlan.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, examplePlan.ID)
		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAuditLogForAccountSubscriptionPlan(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/account_subscription_plans/%d/audit"
		expectedMethod = http.MethodGet
	)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, examplePlan.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(ctx, examplePlan.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(ctx, examplePlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, examplePlan.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForAccountSubscriptionPlan(ctx, examplePlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
