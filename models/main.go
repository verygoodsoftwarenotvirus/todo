package models

import (
	"fmt"
	"strconv"
)

const DefaultLimit = 20

type ContextKey string

type QueryFilter struct {
	Page  uint `json:"page"`
	Limit uint `json:"limit"`
}

func (qf *QueryFilter) ToMap() map[string]string {
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
