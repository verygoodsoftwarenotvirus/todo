package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFleshedOutQueryFilter builds a fully fleshed out QueryFilter.
func BuildFleshedOutQueryFilter() *types.QueryFilter {
	return &types.QueryFilter{
		Page:            10,
		Limit:           20,
		CreatedAfter:    uint64(uint32(fake.Date().Unix())),
		CreatedBefore:   uint64(uint32(fake.Date().Unix())),
		UpdatedAfter:    uint64(uint32(fake.Date().Unix())),
		UpdatedBefore:   uint64(uint32(fake.Date().Unix())),
		SortBy:          types.SortAscending,
		IncludeArchived: true,
	}
}
