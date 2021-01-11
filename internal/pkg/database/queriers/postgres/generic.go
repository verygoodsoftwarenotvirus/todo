package postgres

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

func joinUint64s(in []uint64) string {
	out := []string{}

	for _, x := range in {
		out = append(out, strconv.FormatUint(x, 10))
	}

	return strings.Join(out, ",")
}

// buildQuery builds a given query, handles whatever errors and returns just the query and args.
func (q *Postgres) buildQuery(builder squirrel.Sqlizer) (query string, args []interface{}) {
	var err error

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}
func (q *Postgres) buildTotalCountQuery(tableName, ownershipColumn string, userID uint64, forAdmin, includeArchived bool) (query string, args []interface{}) {
	where := squirrel.Eq{}
	totalCountQueryBuilder := q.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(fmt.Sprintf(columnCountQueryTemplate, tableName)).
		From(tableName)

	if !forAdmin {
		if userID != 0 && ownershipColumn != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		where[fmt.Sprintf("%s.%s", tableName, queriers.ArchivedOnColumn)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.%s", tableName, queriers.ArchivedOnColumn)] = nil
	}

	if len(where) > 0 {
		totalCountQueryBuilder = totalCountQueryBuilder.Where(where)
	}

	return q.buildQuery(totalCountQueryBuilder)
}

func (q *Postgres) buildFilteredCountQuery(tableName, ownershipColumn string, userID uint64, forAdmin, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	where := squirrel.Eq{}
	filteredCountQueryBuilder := q.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(fmt.Sprintf(columnCountQueryTemplate, tableName)).
		From(tableName)

	if !forAdmin {
		if userID != 0 && ownershipColumn != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		where[fmt.Sprintf("%s.%s", tableName, queriers.ArchivedOnColumn)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.%s", tableName, queriers.ArchivedOnColumn)] = nil
	}

	if len(where) > 0 {
		filteredCountQueryBuilder = filteredCountQueryBuilder.Where(where)
	}

	if filter != nil {
		filteredCountQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, filteredCountQueryBuilder, tableName)
	}

	return q.buildQuery(filteredCountQueryBuilder)
}

// buildListQuery builds a SQL query selecting rows that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) buildListQuery(tableName, ownershipColumn string, columns []string, userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
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
			fmt.Sprintf("%s.%s", tableName, queriers.ArchivedOnColumn): nil,
			fmt.Sprintf("%s.%s", tableName, ownershipColumn):           userID,
		})
	}

	builder = builder.GroupBy(fmt.Sprintf("%s.%s", tableName, queriers.IDColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, tableName)
	}

	query, selectArgs := q.buildQuery(builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}
