package requests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func TestBuilder_BuildGetAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccountSubscriptionPlan.ID)

		actual, err := helper.builder.BuildGetAccountSubscriptionPlanRequest(helper.ctx, exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account subscription plan ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAccountSubscriptionPlanRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAccountSubscriptionPlansRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		filter := (*types.QueryFilter)(nil)

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := helper.builder.BuildGetAccountSubscriptionPlansRequest(helper.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildCreateAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildCreateAccountSubscriptionPlanRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateAccountSubscriptionPlanRequest(helper.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateAccountSubscriptionPlanRequest(helper.ctx, &types.AccountSubscriptionPlanCreationInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildUpdateAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleAccountSubscriptionPlan.ID)

		actual, err := helper.builder.BuildUpdateAccountSubscriptionPlanRequest(helper.ctx, exampleAccountSubscriptionPlan)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildUpdateAccountSubscriptionPlanRequest(helper.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildArchiveAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/account_subscription_plans/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccountSubscriptionPlan.ID)

		actual, err := helper.builder.BuildArchiveAccountSubscriptionPlanRequest(helper.ctx, exampleAccountSubscriptionPlan.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid plan ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveAccountSubscriptionPlanRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForAccountSubscriptionPlanRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/account_subscription_plans/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlan()

		actual, err := helper.builder.BuildGetAuditLogForAccountSubscriptionPlanRequest(helper.ctx, exampleAccountSubscriptionPlan.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAccountSubscriptionPlan.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid plan ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAuditLogForAccountSubscriptionPlanRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
