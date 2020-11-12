package dbclient

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhook", mock.Anything, exampleWebhook.ID, exampleWebhook.BelongsToUser).Return(exampleWebhook, nil)

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expected := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetAllWebhooksCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return(exampleWebhookList, nil)

		actual, err := c.GetAllWebhooks(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()
		filter := models.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, exampleUser.ID, filter).Return(exampleWebhookList, nil)

		actual, err := c.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()
		filter := (*models.QueryFilter)(nil)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, exampleUser.ID, filter).Return(exampleWebhookList, nil)

		actual, err := c.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("CreateWebhook", mock.Anything, exampleInput).Return(exampleWebhook, nil)

		actual, err := c.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("UpdateWebhook", mock.Anything, exampleWebhook).Return(expected)

		actual := c.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("ArchiveWebhook", mock.Anything, exampleWebhook.ID, exampleWebhook.BelongsToUser).Return(expected)

		actual := c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
