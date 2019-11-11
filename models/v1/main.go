package models

import (
	"fmt"
)

const (
	// SortAscending is the pre-determined Ascending sortType for external use
	SortAscending sortType = "asc"
	// SortDescending is the pre-determined Descending sortType for external use
	SortDescending sortType = "desc"
)

type (
	// ContextKey represents strings to be used in Context objects. From the docs:
	// 		"The provided key must be comparable and should not be of type string or
	// 		any other built-in type to avoid collisions between packages using context."
	ContextKey string
	sortType   string

	// Pagination represents a pagination request
	Pagination struct {
		Page       uint64 `json:"page"`
		Limit      uint64 `json:"limit"`
		TotalCount uint64 `json:"total_count"`
	}

	// CountResponse is what we respond with when a user requests a count of data types
	CountResponse struct {
		Count uint64 `json:"count"`
	}
)

var _ error = (*ErrorResponse)(nil)

// ErrorResponse represents a response we might send to the user in the event of an error
type ErrorResponse struct {
	Message string `json:"message"`
	Code    uint   `json:"code"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s", er.Code, er.Message)
}
