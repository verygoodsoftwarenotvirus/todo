package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

func joinIDs(ids []uint64) string {
	out := []string{}

	for _, x := range ids {
		out = append(out, strconv.FormatUint(x, 10))
	}

	return strings.Join(out, ",")
}

func joinStringIDs(ids []string) string {
	return strings.Join(ids, ",")
}

// logQueryBuildingError logs errs that may occur during query construction. Such errors should be few and far between,
// as the generally only occur with type discrepancies or other misuses of SQL. An alert should be set up for any log
// entries with the given name, and those alerts should be investigated quickly.
func (b *Postgres) logQueryBuildingError(span tracing.Span, err error) {
	if err != nil {
		logger := b.logger.WithValue(keys.QueryErrorKey, true)
		observability.AcknowledgeError(err, logger, span, "building query")
	}
}

// BuildQueryOnly builds a given query, handles whatever errs and returns just the query and args.
func (b *Postgres) buildQueryOnly(span tracing.Span, builder squirrel.Sqlizer) string {
	query, _, err := builder.ToSql()

	b.logQueryBuildingError(span, err)

	return query
}

// BuildQuery builds a given query, handles whatever errs and returns just the query and args.
func (b *Postgres) buildQuery(span tracing.Span, builder squirrel.Sqlizer) (query string, args []interface{}) {
	query, args, err := builder.ToSql()

	b.logQueryBuildingError(span, err)

	return query, args
}

func (b *Postgres) buildTotalCountQuery(ctx context.Context, tableName string, joins []string, where squirrel.Eq, ownershipColumn, userID string, forAdmin, includeArchived bool) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if where == nil {
		where = squirrel.Eq{}
	}

	totalCountQueryBuilder := b.sqlBuilder.
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

		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	}

	if len(where) > 0 {
		totalCountQueryBuilder = totalCountQueryBuilder.Where(where)
	}

	return b.buildQuery(span, totalCountQueryBuilder)
}

func (b *Postgres) buildFilteredCountQuery(ctx context.Context, tableName string, joins []string, where squirrel.Eq, ownershipColumn, userID string, forAdmin, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	if where == nil {
		where = squirrel.Eq{}
	}

	filteredCountQueryBuilder := b.sqlBuilder.
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

		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	} else if !includeArchived {
		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil
	}

	if len(where) > 0 {
		filteredCountQueryBuilder = filteredCountQueryBuilder.Where(where)
	}

	if filter != nil {
		filteredCountQueryBuilder = querybuilding.ApplyFilterToSubCountQueryBuilder(filter, tableName, filteredCountQueryBuilder)
	}

	return b.buildQuery(span, filteredCountQueryBuilder)
}

// BuildListQuery builds a SQL query selecting rows that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (b *Postgres) buildListQuery(
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
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	var includeArchived bool
	if filter != nil {
		includeArchived = filter.IncludeArchived
	}

	filteredCountQuery, filteredCountQueryArgs := b.buildFilteredCountQuery(ctx, tableName, joins, where, ownershipColumn, ownerID, forAdmin, includeArchived, filter)
	totalCountQuery, totalCountQueryArgs := b.buildTotalCountQuery(ctx, tableName, joins, where, ownershipColumn, ownerID, forAdmin, includeArchived)

	builder := b.sqlBuilder.
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
		where[fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn)] = nil

		if ownershipColumn != "" && ownerID != "" {
			where[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = ownerID
		}

		builder = builder.Where(where)
	}

	builder = builder.GroupBy(fmt.Sprintf("%s.%s", tableName, querybuilding.IDColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, tableName, builder)
	}

	query, selectArgs := b.buildQuery(span, builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}
