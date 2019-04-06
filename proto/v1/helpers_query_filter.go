package todoproto

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// ToModelQueryFilter attaches to QueryFilter to provide itself a way to convert to our models' package version of a QueryFilter
func (qf *QueryFilter) ToModelQueryFilter() *models.QueryFilter {
	var sb = models.SortAscending
	if qf.SortBy == string(models.SortDescending) {
		sb = models.SortAscending
	}

	return &models.QueryFilter{
		Page:          qf.Page,
		Limit:         qf.Limit,
		CreatedAfter:  qf.CreatedAfter,
		CreatedBefore: qf.CreatedBefore,
		UpdatedAfter:  qf.UpdatedAfter,
		UpdatedBefore: qf.UpdatedBefore,
		IncludeAll:    qf.IncludeAll,
		SortBy:        sb,
	}
}
