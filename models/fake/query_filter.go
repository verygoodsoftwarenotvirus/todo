package fakemodels

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFleshedOutQueryFilter builds a fully fleshed out QueryFilter.
func BuildFleshedOutQueryFilter() *models.QueryFilter {
	return &models.QueryFilter{
		Page:          10,
		Limit:         20,
		CreatedAfter:  uint64(uint32(fake.Date().Unix())),
		CreatedBefore: uint64(uint32(fake.Date().Unix())),
		UpdatedAfter:  uint64(uint32(fake.Date().Unix())),
		UpdatedBefore: uint64(uint32(fake.Date().Unix())),
		SortBy:        models.SortAscending,
	}
}
