package mariadb

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestMariaDB_BuildListQuery(T *testing.T) {
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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT column_one, column_two, column_three, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL AND example_table.belongs_to_account = ?) as total_count, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL AND example_table.belongs_to_account = ? AND example_table.created_on > ? AND example_table.created_on < ? AND example_table.last_updated_on > ? AND example_table.last_updated_on < ?) as filtered_count FROM example_table WHERE example_table.archived_on IS NULL AND example_table.belongs_to_account = ? AND example_table.created_on > ? AND example_table.created_on < ? AND example_table.last_updated_on > ? AND example_table.last_updated_on < ? GROUP BY example_table.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.buildListQuery(
			exampleTableName,
			exampleOwnershipColumn,
			exampleColumns,
			exampleUser.ID,
			false,
			filter,
		)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})

	T.Run("for admin without archived", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT column_one, column_two, column_three, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL) as total_count, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.archived_on IS NULL AND example_table.created_on > ? AND example_table.created_on < ? AND example_table.last_updated_on > ? AND example_table.last_updated_on < ?) as filtered_count FROM example_table WHERE example_table.created_on > ? AND example_table.created_on < ? AND example_table.last_updated_on > ? AND example_table.last_updated_on < ? GROUP BY example_table.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.buildListQuery(
			exampleTableName,
			exampleOwnershipColumn,
			exampleColumns,
			exampleUser.ID,
			true,
			filter,
		)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})

	T.Run("for admin with archived", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()
		filter.IncludeArchived = true

		expectedQuery := "SELECT column_one, column_two, column_three, (SELECT COUNT(example_table.id) FROM example_table) as total_count, (SELECT COUNT(example_table.id) FROM example_table WHERE example_table.created_on > ? AND example_table.created_on < ? AND example_table.last_updated_on > ? AND example_table.last_updated_on < ?) as filtered_count FROM example_table WHERE example_table.created_on > ? AND example_table.created_on < ? AND example_table.last_updated_on > ? AND example_table.last_updated_on < ? GROUP BY example_table.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.buildListQuery(
			exampleTableName,
			exampleOwnershipColumn,
			exampleColumns,
			exampleUser.ID,
			true,
			filter,
		)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
