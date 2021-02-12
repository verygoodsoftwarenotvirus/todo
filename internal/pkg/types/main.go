package types

import (
	"fmt"
)

const (
	// SortAscending is the pre-determined Ascending sortType for external use.
	SortAscending sortType = "asc"
	// SortDescending is the pre-determined Descending sortType for external use.
	SortDescending sortType = "desc"
)

type (
	// ContextKey represents strings to be used in Context objects. From the docs:
	// 	"The provided key must be comparable and should not be of type string or
	// 	 any other built-in type to avoid collisions between packages using context."
	ContextKey string
	sortType   string

	// Pagination represents a pagination request.
	Pagination struct {
		Page          uint64 `json:"page"`
		Limit         uint8  `json:"limit"`
		FilteredCount uint64 `json:"filteredCount"`
		TotalCount    uint64 `json:"totalCount"`
	}

	// ErrorResponse represents a response we might send to the User in the event of an error.
	ErrorResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
)

var _ error = (*ErrorResponse)(nil)

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s", er.Code, er.Message)
}
