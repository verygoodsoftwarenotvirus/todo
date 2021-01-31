package sqlite

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// BuildQuery builds a given query, handles whatever errors and returns just the query and args.
func (q *Sqlite) buildQuery(builder squirrel.Sqlizer) (query string, args []interface{}) {
	var err error

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

func (q *Sqlite) buildTotalCountQuery(tableName, ownershipColumn string, userID uint64, forAdmin, includeArchived bool) (query string, args []interface{}) {
	where := squirrel.Eq{}
	totalCountQueryBuilder := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, tableName)).
		From(tableName)

	if !forAdmin {
		if userID != 0 && ownershipColumn != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	}

	if len(where) > 0 {
		totalCountQueryBuilder = totalCountQueryBuilder.Where(where)
	}

	return q.buildQuery(totalCountQueryBuilder)
}

func (q *Sqlite) buildFilteredCountQuery(tableName, ownershipColumn string, userID uint64, forAdmin, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	where := squirrel.Eq{}
	filteredCountQueryBuilder := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, tableName)).
		From(tableName)

	if !forAdmin {
		if userID != 0 && ownershipColumn != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	}

	if len(where) > 0 {
		filteredCountQueryBuilder = filteredCountQueryBuilder.Where(where)
	}

	if filter != nil {
		filteredCountQueryBuilder = querybuilding.ApplyFilterToSubCountQueryBuilder(filter, filteredCountQueryBuilder, tableName)
	}

	return q.buildQuery(filteredCountQueryBuilder)
}

// BuildListQuery builds a SQL query selecting rows that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Sqlite) buildListQuery(tableName, ownershipColumn string, columns []string, userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	var includeArchived bool
	if filter != nil {
		includeArchived = filter.IncludeArchived
	}

	filteredCountQuery, filteredCountQueryArgs := q.buildFilteredCountQuery(
		tableName,
		ownershipColumn,
		userID,
		forAdmin,
		includeArchived,
		filter,
	)

	totalCountQuery, totalCountQueryArgs := q.buildTotalCountQuery(
		tableName,
		ownershipColumn,
		userID,
		forAdmin,
		includeArchived,
	)

	builder := q.sqlBuilder.
		Select(append(
			columns,
			fmt.Sprintf("(%s) as total_count", totalCountQuery),
			fmt.Sprintf("(%s) as filtered_count", filteredCountQuery),
		)...).
		From(tableName)

	if !forAdmin {
		builder = builder.Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn): nil,
			fmt.Sprintf("%s.%s", tableName, ownershipColumn):                userID,
		})
	}

	builder = builder.GroupBy(fmt.Sprintf("%s.%s", tableName, querybuilding.IDColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, builder, tableName)
	}

	query, selectArgs := q.buildQuery(builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}
