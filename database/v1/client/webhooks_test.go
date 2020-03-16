package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleID := fake.Uint64()
		exampleUserID := fake.Uint64()
		expected := &models.Webhook{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhook", mock.Anything, exampleID, exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhook(ctx, exampleID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetWebhookCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := fake.Uint64()
		filter := models.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhookCount", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhookCount(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := fake.Uint64()
		filter := (*models.QueryFilter)(nil)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhookCount", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhookCount(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := fake.Uint64()

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetAllWebhooksCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.WebhookList{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllWebhooks(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := &models.WebhookList{}
		filter := models.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhooks(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := &models.WebhookList{}
		filter := (*models.QueryFilter)(nil)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhooks(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.WebhookCreationInput{}
		expected := &models.Webhook{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("CreateWebhook", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.Webhook{}
		var expected error

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("UpdateWebhook", mock.Anything, exampleInput).Return(expected)

		actual := c.UpdateWebhook(ctx, exampleInput)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleID := fake.Uint64()
		exampleUserID := fake.Uint64()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("ArchiveWebhook", mock.Anything, exampleID, exampleUserID).Return(expected)

		actual := c.ArchiveWebhook(ctx, exampleID, exampleUserID)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}
