package models

import (
	"fmt"
)

const (
	// DefaultLimit represents how many items we return in a response by default.
	DefaultLimit = 20
)

type (
	// ContextKey represents strings to be used in Context objects. From the docs:
	// 		"The provided key must be comparable and should not be of type string or
	//		any other built-in type to avoid collisions between packages using context."
	ContextKey string
	sortType   string
)

var (
	strSortAsc  = "asc"
	strSortDesc = "desc"

	// SortAscending is the pre-determined Ascending sortType for external use
	SortAscending = (sortType)(strSortAsc)
	// SortDescending is the pre-determined Descending sortType for external use
	SortDescending = (sortType)(strSortDesc)
)

// Pagination represents a pagination request
type Pagination struct {
	Page       uint64 `json:"page"`
	Limit      uint64 `json:"limit"`
	TotalCount uint64 `json:"total_count"`
}

var _ error = (*ErrorResponse)(nil)

// ErrorResponse represents a response we might send to the user in the event of an error.
// REFACTORME: Maybe we should just have one response struct that has a nillable error.
type ErrorResponse struct {
	Message string `json:"message"`
	Code    uint   `json:"code"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%d - %s", er.Code, er.Message)
}

// ResponseMetadata is a struct for future use
type ResponseMetadata struct {
	RequestID string         `json:"request_id"`
	Error     *ErrorResponse `json:"error"`
}

// CountResponse is what we respond with when a user requests a count of data types
type CountResponse struct {
	Count uint64 `json:"count"`
}
