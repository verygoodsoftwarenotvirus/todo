package mariadb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestMariaDB_BuildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.id = ? )"
		expectedArgs := []interface{}{
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildItemExistsQuery(exampleItem.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.id = ?"
		expectedArgs := []interface{}{
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildGetItemQuery(exampleItem.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		actualQuery := q.BuildGetAllItemsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_BuildGetBatchOfItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.id > ? AND items.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfItemsQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user, (SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ?) as total_count, (SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ?) as filtered_count FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ? GROUP BY items.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleUser.ID,
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetItemsQuery(exampleUser.ID, false, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetItemsWithIDsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleIDs := []uint64{
			789,
			123,
			456,
		}

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_user FROM items WHERE items.archived_on IS NULL AND items.belongs_to_user = ? AND items.id IN (?,?,?) ORDER BY CASE items.id WHEN 789 THEN 0 WHEN 123 THEN 1 WHEN 456 THEN 2 END LIMIT 20"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleIDs[0],
			exampleIDs[1],
			exampleIDs[2],
		}
		actualQuery, actualArgs := q.BuildGetItemsWithIDsQuery(exampleUser.ID, defaultLimit, exampleIDs, false)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleItem.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO items (external_id,name,details,belongs_to_user) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleItem.ExternalID,
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToUser,
		}
		actualQuery, actualArgs := q.BuildCreateItemQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestMariaDB_BuildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE items SET name = ?, details = ?, last_updated_on = UNIX_TIMESTAMP() WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToUser,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateItemQuery(exampleItem)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID

		expectedQuery := "UPDATE items SET last_updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveItemQuery(exampleItem.ID, exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAuditLogEntriesForItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()

		expectedQuery := fmt.Sprintf("SELECT audit_log.id, audit_log.external_id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE JSON_CONTAINS(audit_log.context, '%d', '$.item_id') ORDER BY audit_log.created_on", exampleItem.ID)
		expectedArgs := []interface{}(nil)
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForItemQuery(exampleItem.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
