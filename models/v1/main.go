package models

import (
	"fmt"
	"net/url"
	"strings"
)

const DefaultLimit = 20

type (
	ContextKey string
	sortType   string
)

var (
	strSortAsc              = "asc"
	strSortDesc             = "desc"
	SortAscending  sortType = (sortType)(strSortAsc)
	SortDescending sortType = (sortType)(strSortDesc)

	typeMap = map[string]func() interface{}{
		"item": func() interface{} { return new(Item) },
	}

	allTypes = []interface{}{
		new(Item),
	}
)

func DetermineTypesOfInterest(params url.Values) []interface{} {
	toc := []interface{}{}
	if x, ok := params["types"]; ok && len(x) > 0 {
		if len(x) == 1 && x[0] == "*" {
			return allTypes
		}

		for _, y := range x {
			if z, ok := typeMap[strings.ToLower(y)]; ok {
				toc = append(toc, z())
			}
		}
	}
	return toc
}

type Pagination struct {
	Page       uint64 `json:"page"`
	Limit      uint64 `json:"limit"`
	TotalCount uint64 `json:"total_count"`
}

var _ error = (*ErrorResponse)(nil)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    uint   `json:"code"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%d - %s", er.Code, er.Message)
}
