package models

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

const DefaultLimit = 20

type ContextKey string

type QueryFilter struct {
	Page  uint `json:"page"`
	Limit uint `json:"limit"`
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
			qf.Page = uint(math.Max(float64(i), 1))
		}
	}

	if limit, ok := params["limit"]; ok && len(limit) >= 1 {
		if i, err := strconv.ParseUint(limit[0], 10, 64); err == nil {
			qf.Limit = uint(math.Max(float64(i), 1))
		}
	}

	// if updatedAfter, ok := params["updated_after"]; ok && len(updatedAfter) >= 1 {
	// 	if i, err := strconv.ParseUint(updatedAfter[0], 10, 64); err == nil {
	// 		qf.UpdatedAfter = time.Unix(int64(i), 0)
	// 	}
	// }
	//
	// if updatedAfter, ok := params["updated_after"]; ok && len(updatedAfter) >= 1 {
	// 	if i, err := strconv.ParseUint(updatedAfter[0], 10, 64); err == nil {
	// 		qf.UpdatedAfter = time.Unix(int64(i), 0)
	// 	}
	// }
	//
	// if updatedAfter, ok := params["updated_after"]; ok && len(updatedAfter) >= 1 {
	// 	if i, err := strconv.ParseUint(updatedAfter[0], 10, 64); err == nil {
	// 		qf.UpdatedAfter = time.Unix(int64(i), 0)
	// 	}
	// }
	//
	// if updatedAfter, ok := params["updated_after"]; ok && len(updatedAfter) >= 1 {
	// 	if i, err := strconv.ParseUint(updatedAfter[0], 10, 64); err == nil {
	// 		qf.UpdatedAfter = time.Unix(int64(i), 0)
	// 	}
	// }
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
