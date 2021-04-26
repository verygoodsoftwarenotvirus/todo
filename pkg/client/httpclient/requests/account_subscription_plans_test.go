package requests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func TestBuilder_BuildGetAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccountSubscriptionPlan.ID)

		actual, err := h.builder.BuildGetAccountSubscriptionPlanRequest(h.ctx, exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account subscription plan ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAccountSubscriptionPlanRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAccountSubscriptionPlansRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		filter := (*types.QueryFilter)(nil)

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := h.builder.BuildGetAccountSubscriptionPlansRequest(h.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildCreateAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := h.builder.BuildCreateAccountSubscriptionPlanRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateAccountSubscriptionPlanRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateAccountSubscriptionPlanRequest(h.ctx, &types.AccountSubscriptionPlanCreationInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildUpdateAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleAccountSubscriptionPlan.ID)

		actual, err := h.builder.BuildUpdateAccountSubscriptionPlanRequest(h.ctx, exampleAccountSubscriptionPlan)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildUpdateAccountSubscriptionPlanRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildArchiveAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccountSubscriptionPlan.ID)

		actual, err := h.builder.BuildArchiveAccountSubscriptionPlanRequest(h.ctx, exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid plan ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildArchiveAccountSubscriptionPlanRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		actual, err := h.builder.BuildGetAuditLogForAccountSubscriptionPlanRequest(h.ctx, exampleAccountSubscriptionPlan.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAccountSubscriptionPlan.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid plan ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAuditLogForAccountSubscriptionPlanRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
