package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
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

func (b *Postgres) buildTotalCountQuery(ctx context.Context, tableName, ownershipColumn string, userID uint64, forAdmin, includeArchived bool) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	where := squirrel.Eq{}
	totalCountQueryBuilder := b.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
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

	return b.buildQuery(span, totalCountQueryBuilder)
}

func (b *Postgres) buildFilteredCountQuery(ctx context.Context, tableName, ownershipColumn string, userID uint64, forAdmin, includeArchived bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}
	where := squirrel.Eq{}
	filteredCountQueryBuilder := b.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
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
		filteredCountQueryBuilder = querybuilding.ApplyFilterToSubCountQueryBuilder(filter, tableName, filteredCountQueryBuilder)
	}

	return b.buildQuery(span, filteredCountQueryBuilder)
}

// BuildListQuery builds a SQL query selecting rows that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (b *Postgres) buildListQuery(ctx context.Context, tableName, ownershipColumn string, columns []string, userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	var includeArchived bool
	if filter != nil {
		includeArchived = filter.IncludeArchived
	}

	filteredCountQuery, filteredCountQueryArgs := b.buildFilteredCountQuery(ctx, tableName, ownershipColumn, userID, forAdmin, includeArchived, filter)
	totalCountQuery, totalCountQueryArgs := b.buildTotalCountQuery(ctx, tableName, ownershipColumn, userID, forAdmin, includeArchived)

	builder := b.sqlBuilder.
		Select(append(
			columns,
			fmt.Sprintf("(%s) as total_count", totalCountQuery),
			fmt.Sprintf("(%s) as filtered_count", filteredCountQuery),
		)...).
		From(tableName)

	if !forAdmin {
		w := squirrel.Eq{fmt.Sprintf("%s.%s", tableName, querybuilding.ArchivedOnColumn): nil}
		if ownershipColumn != "" && userID != 0 {
			w[fmt.Sprintf("%s.%s", tableName, ownershipColumn)] = userID
		}

		builder = builder.Where(w)
	}

	builder = builder.GroupBy(fmt.Sprintf("%s.%s", tableName, querybuilding.IDColumn))

	if filter != nil {
		builder = querybuilding.ApplyFilterToQueryBuilder(filter, tableName, builder)
	}

	query, selectArgs := b.buildQuery(span, builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}
