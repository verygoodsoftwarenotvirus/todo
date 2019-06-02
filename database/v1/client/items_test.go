package dbclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

func TestClient_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleItemID := uint64(123)
		exampleUserID := uint64(123)
		expected := &models.Item{}

		c, db := buildTestClient()
		db.ItemDataManager.
			On("GetItem", mock.Anything, exampleItemID, exampleUserID).
			Return(expected, nil)

		actual, err := c.GetItem(context.Background(), exampleItemID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestClient_GetItemCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		expected := uint64(321)
		exampleUserID := uint64(123)

		c, db := buildTestClient()
		db.ItemDataManager.
			On("GetItemCount", mock.Anything, models.DefaultQueryFilter, exampleUserID).
			Return(expected, nil)

		actual, err := c.GetItemCount(context.Background(), models.DefaultQueryFilter, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestClient_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		expected := uint64(321)
		c, db := buildTestClient()
		db.ItemDataManager.
			On("GetAllItemsCount", mock.Anything).
			Return(expected, nil)

		actual, err := c.GetAllItemsCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestClient_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleUserID := uint64(123)
		c, db := buildTestClient()
		expected := &models.ItemList{}

		db.ItemDataManager.
			On("GetItems", mock.Anything, models.DefaultQueryFilter, exampleUserID).
			Return(expected, nil)

		actual, err := c.GetItems(context.Background(), models.DefaultQueryFilter, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestClient_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := &models.ItemInput{}
		c, db := buildTestClient()
		expected := &models.Item{}

		db.ItemDataManager.
			On("CreateItem", mock.Anything, exampleInput).
			Return(expected, nil)

		actual, err := c.CreateItem(context.Background(), exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestClient_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := &models.Item{}
		c, db := buildTestClient()
		var expected error

		db.ItemDataManager.
			On("UpdateItem", mock.Anything, exampleInput).
			Return(expected)

		err := c.UpdateItem(context.Background(), exampleInput)
		assert.NoError(t, err)
	})
}

func TestClient_DeleteItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleUserID := uint64(123)
		exampleItemID := uint64(123)
		var expected error

		c, db := buildTestClient()
		db.ItemDataManager.On("DeleteItem", mock.Anything, exampleItemID, exampleUserID).
			Return(expected)

		err := c.DeleteItem(context.Background(), exampleUserID, exampleItemID)
		assert.NoError(t, err)
	})
}
