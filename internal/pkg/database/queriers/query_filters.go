package queriers

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// ApplyFilterToQueryBuilder applies the query filter to a query builder.
func ApplyFilterToQueryBuilder(qf *types.QueryFilter, queryBuilder squirrel.SelectBuilder, tableName string) squirrel.SelectBuilder {
	if qf == nil {
		return queryBuilder
	}

	qf.SetPage(qf.Page)

	if qp := qf.QueryPage(); qp > 0 {
		queryBuilder = queryBuilder.Offset(qp)
	}

	if qf.Limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(qf.Limit))
	} else {
		queryBuilder = queryBuilder.Limit(types.MaxLimit)
	}

	if qf.CreatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, CreatedOnColumn): qf.CreatedAfter})
	}

	if qf.CreatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, CreatedOnColumn): qf.CreatedBefore})
	}

	if qf.UpdatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, LastUpdatedOnColumn): qf.UpdatedAfter})
	}

	if qf.UpdatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, LastUpdatedOnColumn): qf.UpdatedBefore})
	}

	return queryBuilder
}

// ApplyFilterToSubCountQueryBuilder applies the query filter to a query builder.
func ApplyFilterToSubCountQueryBuilder(qf *types.QueryFilter, queryBuilder squirrel.SelectBuilder, tableName string) squirrel.SelectBuilder {
	if qf == nil {
		return queryBuilder
	}

	if qf.CreatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, CreatedOnColumn): qf.CreatedAfter})
	}

	if qf.CreatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, CreatedOnColumn): qf.CreatedBefore})
	}

	if qf.UpdatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, LastUpdatedOnColumn): qf.UpdatedAfter})
	}

	if qf.UpdatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, LastUpdatedOnColumn): qf.UpdatedBefore})
	}

	if qf.IncludeArchived {
		queryBuilder = queryBuilder.Where(squirrel.Or{
			squirrel.Eq{fmt.Sprintf("%s.%s", tableName, ArchivedOnColumn): nil},
			squirrel.NotEq{fmt.Sprintf("%s.%s", tableName, ArchivedOnColumn): nil},
		})
	}

	return queryBuilder
}
