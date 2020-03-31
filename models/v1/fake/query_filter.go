package fake

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit"
)

// BuildFleshedOutQueryFilter builds a fully fleshed out QueryFilter
func BuildFleshedOutQueryFilter() *models.QueryFilter {
	return &models.QueryFilter{
		Page:          10,
		Limit:         20,
		CreatedAfter:  uint64(fake.Date().Unix()),
		CreatedBefore: uint64(fake.Date().Unix()),
		UpdatedAfter:  uint64(fake.Date().Unix()),
		UpdatedBefore: uint64(fake.Date().Unix()),
		SortBy:        models.SortAscending,
	}
}
