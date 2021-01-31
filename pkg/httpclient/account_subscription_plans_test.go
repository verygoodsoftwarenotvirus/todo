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

func TestV1Client_BuildGetPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, examplePlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetPlanRequest(ctx, examplePlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetPlan(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/plans/%d"

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
		actual, err := c.GetPlan(ctx, examplePlan.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, examplePlan, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetPlan(ctx, examplePlan.ID)

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
		actual, err := c.GetPlan(ctx, examplePlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetPlansRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/plans"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetPlansRequest(ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetPlans(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/plans"

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
		actual, err := c.GetPlans(ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, examplePlanList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetPlans(ctx, filter)

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
		actual, err := c.GetPlans(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildCreatePlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/plans"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(examplePlan)

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildCreatePlanRequest(ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_CreatePlan(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/plans"

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
		actual, err := c.CreatePlan(ctx, exampleInput)

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
		actual, err := c.CreatePlan(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildUpdatePlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, examplePlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdatePlanRequest(ctx, examplePlan)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_UpdatePlan(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/plans/%d"

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

		err := buildTestClient(t, ts).UpdatePlan(ctx, examplePlan)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		err := buildTestClientWithInvalidURL(t).UpdatePlan(ctx, examplePlan)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildArchivePlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/plans/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, examplePlan.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildArchivePlanRequest(ctx, examplePlan.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_ArchivePlan(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/plans/%d"

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

		err := buildTestClient(t, ts).ArchivePlan(ctx, examplePlan.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		err := buildTestClientWithInvalidURL(t).ArchivePlan(ctx, examplePlan.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetAuditLogForPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/plans/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForPlanRequest(ctx, examplePlan.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, examplePlan.ID)
		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetAuditLogForPlan(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/plans/%d/audit"
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
		actual, err := c.GetAuditLogForPlan(ctx, examplePlan.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakeAccountSubscriptionPlan()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForPlan(ctx, examplePlan.ID)

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
		actual, err := c.GetAuditLogForPlan(ctx, examplePlan.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
