package superclient

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetPlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakePlan()

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(examplePlan, nil)

		actual, err := c.GetAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllPlansCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountSubscriptionPlanDataManager.On("GetAllAccountSubscriptionPlansCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetAllAccountSubscriptionPlansCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetPlans(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		examplePlanList := fakes.BuildFakePlanList()

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.Anything, filter).Return(examplePlanList, nil)

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, examplePlanList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		examplePlanList := fakes.BuildFakePlanList()

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.Anything, filter).Return(examplePlanList, nil)

		actual, err := c.GetAccountSubscriptionPlans(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, examplePlanList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreatePlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountSubscriptionPlanDataManager.On("CreateAccountSubscriptionPlan", mock.Anything, exampleInput).Return(examplePlan, nil)

		actual, err := c.CreateAccountSubscriptionPlan(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, examplePlan, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdatePlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		var expected error

		examplePlan := fakes.BuildFakePlan()

		c, _, mockDB := buildTestClient(t)

		mockDB.AccountSubscriptionPlanDataManager.On("UpdateAccountSubscriptionPlan", mock.Anything, examplePlan).Return(expected)

		err := c.UpdateAccountSubscriptionPlan(ctx, examplePlan)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchivePlan(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		var expected error

		examplePlan := fakes.BuildFakePlan()

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(expected)

		err := c.ArchiveAccountSubscriptionPlan(ctx, examplePlan.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
