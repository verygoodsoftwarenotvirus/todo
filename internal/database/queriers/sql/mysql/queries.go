package postgres

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

const (
	commaSeparator           = ","
	columnCountQueryTemplate = `COUNT(%s.id)`
	userOwnershipColumn      = "belongs_to_user"
	accountOwnershipColumn   = "belongs_to_account"
)

func joinIDs(ids []string) string {
	return strings.Join(ids, commaSeparator)
}

// logQueryBuildingError logs errs that may occur during query construction. Such errors should be few and far between,
// as the generally only occur with type discrepancies or other misuses of SQL. An alert should be set up for any log
// entries with the given name, and those alerts should be investigated quickly.
func (q *SQLQuerier) logQueryBuildingError(span tracing.Span, err error) {
	if err != nil {
		logger := q.logger.WithValue(keys.QueryErrorKey, true)
		observability.AcknowledgeError(err, logger, span, "building query")
	}
}

// BuildQuery builds a given query, handles whatever errs and returns just the query and args.
func (q *SQLQuerier) buildQuery(span tracing.Span, builder squirrel.Sqlizer) (query string, args []interface{}) {
	query, args, err := builder.ToSql()

	q.logQueryBuildingError(span, err)

	return query, args
}

func (q *SQLQuerier) buildTotalCountQuery(ctx context.Context, tableName string, joins []string, where squirrel.Eq, ownershipColumn, userID string, forAdmin, includeArchived bool) (query string, args []interface{}) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if where == nil {
		where = squirrel.Eq{}
	}

	totalCountQueryBuilder := q.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(fmt.Sprintf(columnCountQueryTemplate, tableName)).
		From(tableName)

	for _, join := range joins {
		totalCountQueryBuilder = totalCountQueryBuilder.Join(join)
	}

	if !forAdmin {
		if userID != "" && ownershipColumn != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		where[fmt.Sprintf("%s.archived_on", tableName)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.archived_on", tableName)] = nil
	}

	if len(where) > 0 {
		totalCountQueryBuilder = totalCountQueryBuilder.Where(where)
	}

	return q.buildQuery(span, totalCountQueryBuilder)
}

func (q *SQLQuerier) buildFilteredCountQuery(ctx context.Context, tableName string, joins []string, where squirrel.Eq, ownershipColumn, userID string, forAdmin, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	if where == nil {
		where = squirrel.Eq{}
	}

	filteredCountQueryBuilder := q.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(fmt.Sprintf(columnCountQueryTemplate, tableName)).
		From(tableName)

	for _, join := range joins {
		filteredCountQueryBuilder = filteredCountQueryBuilder.Join(join)
	}

	if !forAdmin {
		if userID != "" && ownershipColumn != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		where[fmt.Sprintf("%s.archived_on", tableName)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.archived_on", tableName)] = nil
	}

	if len(where) > 0 {
		filteredCountQueryBuilder = filteredCountQueryBuilder.Where(where)
	}

	if filter != nil {
		filteredCountQueryBuilder = applyFilterToSubCountQueryBuilder(filter, tableName, filteredCountQueryBuilder)
	}

	return q.buildQuery(span, filteredCountQueryBuilder)
}

// BuildListQuery builds a SQL query selecting rows that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (q *SQLQuerier) buildListQuery(
	ctx context.Context,
	tableName string,
	joins []string,
	where squirrel.Eq,
	ownershipColumn string,
	columns []string,
	ownerID string,
	forAdmin bool,
	filter *types.QueryFilter,
) (query string, args []interface{}) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	var includeArchived bool
	if filter != nil {
		includeArchived = filter.IncludeArchived
	}

	filteredCountQuery, filteredCountQueryArgs := q.buildFilteredCountQuery(ctx, tableName, joins, where, ownershipColumn, ownerID, forAdmin, includeArchived, filter)
	totalCountQuery, totalCountQueryArgs := q.buildTotalCountQuery(ctx, tableName, joins, where, ownershipColumn, ownerID, forAdmin, includeArchived)

	builder := q.sqlBuilder.
		Select(append(
			columns,
			fmt.Sprintf("(%s) as total_count", totalCountQuery),
			fmt.Sprintf("(%s) as filtered_count", filteredCountQuery),
		)...).
		From(tableName)

	for _, join := range joins {
		builder = builder.Join(join)
	}

	if !forAdmin {
		if where == nil {
			where = squirrel.Eq{}
		}
		where[fmt.Sprintf("%s.archived_on", tableName)] = nil

		if ownershipColumn != "" && ownerID != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = ownerID
		}

		builder = builder.Where(where)
	}

	builder = builder.GroupBy(fmt.Sprintf("%s.%s", tableName, "id"))

	if filter != nil {
		builder = applyFilterToQueryBuilder(filter, tableName, builder)
	}

	query, selectArgs := q.buildQuery(span, builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}
