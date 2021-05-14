package postgres

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildListQuery(T *testing.T) {
	T.Parallel()

	const (
		exampleTableName       = "example_table"
		exampleOwnershipColumn = "belongs_to_account"
	)

	exampleColumns := []string{
		"column_one",
		"column_two",
		"column_three",
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT column_one, column_two, column_three, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL AND example_table.belongs_to_account = $1) as total_count, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL AND example_table.belongs_to_account = $2 AND example_table.created_on > $3 AND example_table.created_on < $4 AND example_table.last_updated_on > $5 AND example_table.last_updated_on < $6) as filtered_count FROM example_table WHERE example_table.archived_on IS NULL AND example_table.belongs_to_account = $7 AND example_table.created_on > $8 AND example_table.created_on < $9 AND example_table.last_updated_on > $10 AND example_table.last_updated_on < $11 GROUP BY example_table.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.buildListQuery(ctx, exampleTableName, exampleOwnershipColumn, exampleColumns, exampleUser.ID, false, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})

	T.Run("for admin without archived", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT column_one, column_two, column_three, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL) as total_count, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL AND example_table.created_on > $1 AND example_table.created_on < $2 AND example_table.last_updated_on > $3 AND example_table.last_updated_on < $4) as filtered_count FROM example_table WHERE example_table.created_on > $5 AND example_table.created_on < $6 AND example_table.last_updated_on > $7 AND example_table.last_updated_on < $8 GROUP BY example_table.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.buildListQuery(ctx, exampleTableName, exampleOwnershipColumn, exampleColumns, exampleUser.ID, true, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})

	T.Run("for admin with archived", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()
		filter.IncludeArchived = true

		expectedQuery := "SELECT column_one, column_two, column_three, (SELECT COUNT(example_table.id) FROM example_table) as total_count, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.created_on > $1 AND example_table.created_on < $2 AND example_table.last_updated_on > $3 AND example_table.last_updated_on < $4) as filtered_count FROM example_table WHERE example_table.created_on > $5 AND example_table.created_on < $6 AND example_table.last_updated_on > $7 AND example_table.last_updated_on < $8 GROUP BY example_table.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.buildListQuery(ctx, exampleTableName, exampleOwnershipColumn, exampleColumns, exampleUser.ID, true, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
