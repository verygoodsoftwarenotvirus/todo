package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleItemID := uint64(123)
		exampleUserID := uint64(123)
		expected := &models.Item{}

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItem", mock.Anything, exampleItemID, exampleUserID).Return(expected, nil)

		actual, err := c.GetItem(context.Background(), exampleItemID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		expected := uint64(321)
		exampleUserID := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItemCount", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, nil)

		actual, err := c.GetItemCount(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		expected := uint64(321)
		exampleUserID := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItemCount", mock.Anything, (*models.QueryFilter)(nil), exampleUserID).Return(expected, nil)

		actual, err := c.GetItemCount(context.Background(), nil, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		expected := uint64(321)
		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetAllItemsCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllItemsCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleUserID := uint64(123)
		c, mockDB := buildTestClient()
		expected := &models.ItemList{}

		mockDB.ItemDataManager.On("GetItems", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, nil)

		actual, err := c.GetItems(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		exampleUserID := uint64(123)
		c, mockDB := buildTestClient()
		expected := &models.ItemList{}

		mockDB.ItemDataManager.On("GetItems", mock.Anything, (*models.QueryFilter)(nil), exampleUserID).Return(expected, nil)

		actual, err := c.GetItems(context.Background(), nil, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := &models.ItemCreationInput{}
		c, mockDB := buildTestClient()
		expected := &models.Item{}

		mockDB.ItemDataManager.On("CreateItem", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateItem(context.Background(), exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := &models.Item{}
		c, mockDB := buildTestClient()
		var expected error

		mockDB.ItemDataManager.On("UpdateItem", mock.Anything, exampleInput).Return(expected)

		err := c.UpdateItem(context.Background(), exampleInput)
		assert.NoError(t, err)
	})
}

func TestClient_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleUserID := uint64(123)
		exampleItemID := uint64(123)
		var expected error

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("ArchiveItem", mock.Anything, exampleItemID, exampleUserID).Return(expected)

		err := c.ArchiveItem(context.Background(), exampleUserID, exampleItemID)
		assert.NoError(t, err)
	})
}
