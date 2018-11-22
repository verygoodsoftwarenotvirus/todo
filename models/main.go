package models

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
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

type QueryFilter struct {
	Page          uint64   `json:"page"`
	Limit         uint64   `json:"limit"`
	CreatedAfter  uint64   `json:"created_before"`
	CreatedBefore uint64   `json:"created_after"`
	UpdatedAfter  uint64   `json:"updated_before"`
	UpdatedBefore uint64   `json:"updated_after"`
	SortBy        sortType `json:"sort_by"`
}

func buildDefaultQueryFilter() *QueryFilter {
	return &QueryFilter{
		Page:  1,
		Limit: DefaultLimit,
	}
}

func (qf *QueryFilter) FromParams(params url.Values) {
	if page, ok := params["page"]; ok && len(page) >= 1 {
		if i, err := strconv.ParseUint(page[0], 10, 64); err == nil {
			qf.Page = uint64(math.Max(float64(i), 1))
		}
	}

	if limit, ok := params["limit"]; ok && len(limit) >= 1 {
		if i, err := strconv.ParseUint(limit[0], 10, 64); err == nil {
			qf.Limit = uint64(math.Max(float64(i), 1))
		}
	}

	if createdBefore, ok := params["created_before"]; ok && len(createdBefore) >= 1 {
		if i, err := strconv.ParseUint(createdBefore[0], 10, 64); err == nil {
			qf.CreatedBefore = uint64(i)
		}
	}

	if createdAfter, ok := params["created_after"]; ok && len(createdAfter) >= 1 {
		if i, err := strconv.ParseUint(createdAfter[0], 10, 64); err == nil {
			qf.CreatedAfter = uint64(i)
		}
	}

	if updatedBefore, ok := params["updated_before"]; ok && len(updatedBefore) >= 1 {
		if i, err := strconv.ParseUint(updatedBefore[0], 10, 64); err == nil {
			qf.UpdatedAfter = uint64(i)
		}
	}

	if updatedAfter, ok := params["updated_after"]; ok && len(updatedAfter) >= 1 {
		if i, err := strconv.ParseUint(updatedAfter[0], 10, 64); err == nil {
			qf.UpdatedAfter = uint64(i)
		}
	}

	if sortBy, ok := params["sort_by"]; ok && len(sortBy) >= 1 {
		switch strings.ToLower(sortBy[0]) {
		case strSortAsc:
			qf.SortBy = SortAscending
		case strSortDesc:
			qf.SortBy = SortDescending
		}
	}
}

func ParseQueryFilter(req *http.Request) *QueryFilter {
	qf := buildDefaultQueryFilter()
	qf.FromParams(req.URL.Query())
	return qf
}

func (qf *QueryFilter) ToMap() map[string]string {
	if qf == nil {
		return DefaultQueryFilter.ToMap()
	}

	return map[string]string{
		"page":  strconv.Itoa(int(qf.Page)),
		"limit": strconv.Itoa(int(qf.Limit)),
	}
}

var DefaultQueryFilter = &QueryFilter{
	Page:  1,
	Limit: DefaultLimit,
}

var _ error = (*ErrorResponse)(nil)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    uint   `json:"code"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%d - %s", er.Code, er.Message)
}
