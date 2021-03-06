package fakes

import (
	fake "github.com/brianvoe/gofakeit/v5"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// BuildFleshedOutQueryFilter builds a fully fleshed out QueryFilter.
func BuildFleshedOutQueryFilter() *types.QueryFilter {
	return &types.QueryFilter{
		Page:          10,
		Limit:         20,
		CreatedAfter:  uint64(uint32(fake.Date().Unix())),
		CreatedBefore: uint64(uint32(fake.Date().Unix())),
		UpdatedAfter:  uint64(uint32(fake.Date().Unix())),
		UpdatedBefore: uint64(uint32(fake.Date().Unix())),
		SortBy:        types.SortAscending,
	}
}
