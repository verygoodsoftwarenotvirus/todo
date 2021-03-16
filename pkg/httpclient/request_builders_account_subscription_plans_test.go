package httpclient

import (
	"context"
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
