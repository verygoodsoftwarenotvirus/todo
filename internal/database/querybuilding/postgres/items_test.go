package postgres

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostgres_BuildItemExistsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		expectedQuery := "SELECT EXISTS ( SELECT items.id FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1 AND items.id = $2 )"
		expectedArgs := []interface{}{
			exampleItem.BelongsToAccount,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildItemExistsQuery(ctx, exampleItem.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1 AND items.id = $2"
		expectedArgs := []interface{}{
			exampleItem.BelongsToAccount,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildGetItemQuery(ctx, exampleItem.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllItemsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"
		actualQuery := q.BuildGetAllItemsCountQuery(ctx)

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_BuildGetBatchOfItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM items WHERE items.id > $1 AND items.id < $2"
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

func TestPostgres_BuildGetItemsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account, (SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1) as total_count, (SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $2 AND items.created_on > $3 AND items.created_on < $4 AND items.last_updated_on > $5 AND items.last_updated_on < $6) as filtered_count FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $7 AND items.created_on > $8 AND items.created_on < $9 AND items.last_updated_on > $10 AND items.last_updated_on < $11 GROUP BY items.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetItemsQuery(ctx, exampleUser.ID, false, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetItemsWithIDsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleIDs := []uint64{
			789,
			123,
			456,
		}
		exampleIDsAsStrings := joinIDsForQuery(exampleIDs)

		expectedQuery := fmt.Sprintf("SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM (SELECT items.id, items.external_id, items.name, items.details, items.created_on, items.last_updated_on, items.archived_on, items.belongs_to_account FROM items JOIN unnest('{%s}'::int[]) WITH ORDINALITY t(id, ord) USING (id) ORDER BY t.ord LIMIT %d) AS items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1", exampleIDsAsStrings, defaultLimit)
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.BuildGetItemsWithIDsQuery(ctx, exampleUser.ID, defaultLimit, exampleIDs, false)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildCreateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleItem.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO items (external_id,name,details,belongs_to_account) VALUES ($1,$2,$3,$4) RETURNING id"
		expectedArgs := []interface{}{
			exampleItem.ExternalID,
			exampleItem.Name,
			exampleItem.Details,
			exampleItem.BelongsToAccount,
		}
		actualQuery, actualArgs := q.BuildCreateItemQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestPostgres_BuildUpdateItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		expectedQuery := "UPDATE items SET name = $1, details = $2, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $3 AND id = $4"
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

func TestPostgres_BuildArchiveItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID

		expectedQuery := "UPDATE items SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $1 AND id = $2"
		expectedArgs := []interface{}{
			exampleAccount.ID,
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveItemQuery(ctx, exampleItem.ID, exampleAccount.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAuditLogEntriesForItemQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		expectedQuery := "SELECT audit_log.id, audit_log.external_id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE audit_log.context->'item_id' = $1 ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			exampleItem.ID,
		}
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForItemQuery(ctx, exampleItem.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
