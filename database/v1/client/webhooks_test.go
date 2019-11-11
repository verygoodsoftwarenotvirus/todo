package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleID := uint64(123)
		exampleUserID := uint64(321)
		expected := &models.Webhook{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhook", mock.Anything, exampleID, exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhook(context.Background(), exampleID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetWebhookCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(321)
		expected := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhookCount", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhookCount(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		exampleUserID := uint64(321)
		expected := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhookCount", mock.Anything, (*models.QueryFilter)(nil), exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhookCount(context.Background(), nil, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetAllWebhooksCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllWebhooksCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.WebhookList{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllWebhooks(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(321)
		expected := &models.WebhookList{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		exampleUserID := uint64(321)
		expected := &models.WebhookList{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, (*models.QueryFilter)(nil), exampleUserID).Return(expected, nil)

		actual, err := c.GetWebhooks(context.Background(), nil, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleInput := &models.WebhookCreationInput{}
		expected := &models.Webhook{}

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("CreateWebhook", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateWebhook(context.Background(), exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleInput := &models.Webhook{}
		var expected error

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("UpdateWebhook", mock.Anything, exampleInput).Return(expected)

		actual := c.UpdateWebhook(context.Background(), exampleInput)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleID := uint64(123)
		exampleUserID := uint64(321)
		var expected error

		c, mockDB := buildTestClient()
		mockDB.WebhookDataManager.On("ArchiveWebhook", mock.Anything, exampleID, exampleUserID).Return(expected)

		actual := c.ArchiveWebhook(context.Background(), exampleID, exampleUserID)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}
