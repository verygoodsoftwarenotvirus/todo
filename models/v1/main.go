package models

import (
	"fmt"
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

type ResponseMetadata struct {
	RequestID string         `json:"request_id"`
	Error     *ErrorResponse `json:"error"`
}

type CountResponse struct {
	Count uint64 `json:"count"`
}
