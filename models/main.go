package models

import (
	"fmt"
)

const DefaultLimit = 20

type ContextKey string

type QueryFilter struct {
	Page  uint `json:"page"`
	Limit uint `json:"limit"`
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
