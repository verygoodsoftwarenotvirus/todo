package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_ItemExists(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleItem.BelongsToUser).Return(true, nil)

		actual, err := c.ItemExists(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleItem.BelongsToUser).Return(exampleItem, nil)

		actual, err := c.GetItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllItemsCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetAllItemsCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetAllItemsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllItems(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []models.Item)

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetAllItems", mock.Anything, results).Return(nil)

		err := c.GetAllItems(ctx, results)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetItems(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()
		exampleItemList := fakemodels.BuildFakeItemList()

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItems", mock.Anything, exampleUser.ID, filter).Return(exampleItemList, nil)

		actual, err := c.GetItems(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*models.QueryFilter)(nil)
		exampleItemList := fakemodels.BuildFakeItemList()

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItems", mock.Anything, exampleUser.ID, filter).Return(exampleItemList, nil)

		actual, err := c.GetItems(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetItemsWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItemList := fakemodels.BuildFakeItemList().Items
		var exampleIDs []uint64
		for _, x := range exampleItemList {
			exampleIDs = append(exampleIDs, x.ID)
		}

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, defaultLimit, exampleIDs).Return(exampleItemList, nil)

		actual, err := c.GetItemsWithIDs(ctx, exampleUser.ID, defaultLimit, exampleIDs)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("CreateItem", mock.Anything, exampleInput).Return(exampleItem, nil)

		actual, err := c.CreateItem(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleItem, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		var expected error

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		c, mockDB := buildTestClient()

		mockDB.ItemDataManager.On("UpdateItem", mock.Anything, exampleItem).Return(expected)

		err := c.UpdateItem(ctx, exampleItem)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		var expected error

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		c, mockDB := buildTestClient()
		mockDB.ItemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleItem.BelongsToUser).Return(expected)

		err := c.ArchiveItem(ctx, exampleItem.ID, exampleItem.BelongsToUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
