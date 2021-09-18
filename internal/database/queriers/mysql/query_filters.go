package mysql

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

// applyFilterToQueryBuilder applies the query filter to a query builder.
func applyFilterToQueryBuilder(qf *types.QueryFilter, tableName string, queryBuilder squirrel.SelectBuilder) squirrel.SelectBuilder {
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
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, "created_on"): qf.CreatedAfter})
	}

	if qf.CreatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, "created_on"): qf.CreatedBefore})
	}

	if qf.UpdatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, "last_updated_on"): qf.UpdatedAfter})
	}

	if qf.UpdatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, "last_updated_on"): qf.UpdatedBefore})
	}

	return queryBuilder
}

// applyFilterToSubCountQueryBuilder applies the query filter to a query builder.
func applyFilterToSubCountQueryBuilder(qf *types.QueryFilter, tableName string, queryBuilder squirrel.SelectBuilder) squirrel.SelectBuilder {
	if qf == nil {
		return queryBuilder
	}

	if qf.CreatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, "created_on"): qf.CreatedAfter})
	}

	if qf.CreatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, "created_on"): qf.CreatedBefore})
	}

	if qf.UpdatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, "last_updated_on"): qf.UpdatedAfter})
	}

	if qf.UpdatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, "last_updated_on"): qf.UpdatedBefore})
	}

	return queryBuilder
}
