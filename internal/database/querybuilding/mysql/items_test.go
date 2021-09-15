package mysql

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestMySQL_BuildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = ? AND items.id = ? )"
		expectedArgs := []interface{}{
			exampleAccountID,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildItemExistsQuery(ctx, exampleItem.ID, exampleAccountID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleItem := fakes.BuildFakeItem()

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = ? AND items.id = ?"
		expectedArgs := []interface{}{
			exampleAccountID,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildGetItemQuery(ctx, exampleItem.ID, exampleAccountID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildGetTotalItemCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		actualQuery := q.BuildGetTotalItemCountQuery(ctx)

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMySQL_BuildGetBatchOfItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM items WHERE items.id > ? AND items.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfItemsQuery(ctx, beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account, (SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = ?) as total_count, (SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ?) as filtered_count FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ? GROUP BY items.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleAccountID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleAccountID,
			exampleAccountID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetItemsQuery(ctx, exampleAccountID, false, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildGetItemsWithIDsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleIDs := []string{
			"789",
			"123",
			"456",
		}

		expectedQuery := "SELECT items.id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = ? AND items.id IN (?,?,?) ORDER BY CASE items.id WHEN '789' THEN 0 WHEN '123' THEN 1 WHEN '456' THEN 2 END LIMIT 20"
		expectedArgs := []interface{}{
			exampleAccountID,
			exampleIDs[0],
			exampleIDs[1],
			exampleIDs[2],
		}
		actualQuery, actualArgs := q.BuildGetItemsWithIDsQuery(ctx, exampleAccountID, defaultLimit, exampleIDs, true)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)

		expectedQuery := "INSERT INTO items (id,name,details,belongs_to_account) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleItem.ID,
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToAccount,
		}
		actualQuery, actualArgs := q.BuildCreateItemQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		expectedQuery := "UPDATE items SET name = ?, details = ?, last_updated_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToAccount,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateItemQuery(ctx, exampleItem)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMySQL_BuildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccountID := fakes.BuildFakeID()
		exampleItemID := fakes.BuildFakeID()

		expectedQuery := "UPDATE items SET last_updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleAccountID,
			exampleItemID,
		}
		actualQuery, actualArgs := q.BuildArchiveItemQuery(ctx, exampleItemID, exampleAccountID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
