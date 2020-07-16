package models

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
)

const (
	// MaxLimit is the maximum value for list queries.
	MaxLimit = 250
	// DefaultLimit represents how many results we return in a response by default.
	DefaultLimit = 20

	// SearchQueryKey is the query param key we use to find search queries in requests
	SearchQueryKey = "q"
	// LimitQueryKey is the query param key we use to specify a limit in a query
	LimitQueryKey = "limit"

	pageQueryKey          = "page"
	createdBeforeQueryKey = "createdBefore"
	createdAfterQueryKey  = "createdAfter"
	updatedBeforeQueryKey = "updatedBefore"
	updatedAfterQueryKey  = "updatedAfter"
	sortByQueryKey        = "sortBy"
)

// QueryFilter represents all the filters a user could apply to a list query.
type QueryFilter struct {
	Page          uint64   `json:"page"`
	Limit         uint8    `json:"limit"`
	CreatedAfter  uint64   `json:"createdBefore,omitempty"`
	CreatedBefore uint64   `json:"createdAfter,omitempty"`
	UpdatedAfter  uint64   `json:"updatedBefore,omitempty"`
	UpdatedBefore uint64   `json:"updatedAfter,omitempty"`
	SortBy        sortType `json:"sortBy"`
}

// DefaultQueryFilter builds the default query filter.
func DefaultQueryFilter() *QueryFilter {
	return &QueryFilter{
		Page:   1,
		Limit:  DefaultLimit,
		SortBy: SortAscending,
	}
}

// FromParams overrides the core QueryFilter values with values retrieved from url.Params
func (qf *QueryFilter) FromParams(params url.Values) {
	if i, err := strconv.ParseUint(params.Get(pageQueryKey), 10, 64); err == nil {
		qf.Page = uint64(math.Max(float64(i), 1))
	}

	if i, err := strconv.ParseUint(params.Get(LimitQueryKey), 10, 64); err == nil {
		qf.Limit = uint8(math.Min(math.Max(float64(i), 0), MaxLimit))
	}

	if i, err := strconv.ParseUint(params.Get(createdBeforeQueryKey), 10, 64); err == nil {
		qf.CreatedBefore = uint64(math.Max(float64(i), 0))
	}

	if i, err := strconv.ParseUint(params.Get(createdAfterQueryKey), 10, 64); err == nil {
		qf.CreatedAfter = uint64(math.Max(float64(i), 0))
	}

	if i, err := strconv.ParseUint(params.Get(updatedBeforeQueryKey), 10, 64); err == nil {
		qf.UpdatedBefore = uint64(math.Max(float64(i), 0))
	}

	if i, err := strconv.ParseUint(params.Get(updatedAfterQueryKey), 10, 64); err == nil {
		qf.UpdatedAfter = uint64(math.Max(float64(i), 0))
	}

	switch strings.ToLower(params.Get(sortByQueryKey)) {
	case string(SortAscending):
		qf.SortBy = SortAscending
	case string(SortDescending):
		qf.SortBy = SortDescending
	}
}

// SetPage sets the current page with certain constraints.
func (qf *QueryFilter) SetPage(page uint64) {
	qf.Page = uint64(math.Max(1, float64(page)))
}

// QueryPage calculates a query page from the current filter values.
func (qf *QueryFilter) QueryPage() uint64 {
	return uint64(qf.Limit) * (qf.Page - 1)
}

// ToValues returns a url.Values from a QueryFilter
func (qf *QueryFilter) ToValues() url.Values {
	if qf == nil {
		return DefaultQueryFilter().ToValues()
	}

	v := url.Values{}
	if qf.Page != 0 {
		v.Set(pageQueryKey, strconv.FormatUint(qf.Page, 10))
	}
	if qf.Limit != 0 {
		v.Set(LimitQueryKey, strconv.FormatUint(uint64(qf.Limit), 10))
	}
	if qf.SortBy != "" {
		v.Set(sortByQueryKey, string(qf.SortBy))
	}
	if qf.CreatedBefore != 0 {
		v.Set(createdBeforeQueryKey, strconv.FormatUint(qf.CreatedBefore, 10))
	}
	if qf.CreatedAfter != 0 {
		v.Set(createdAfterQueryKey, strconv.FormatUint(qf.CreatedAfter, 10))
	}
	if qf.UpdatedBefore != 0 {
		v.Set(updatedBeforeQueryKey, strconv.FormatUint(qf.UpdatedBefore, 10))
	}
	if qf.UpdatedAfter != 0 {
		v.Set(updatedAfterQueryKey, strconv.FormatUint(qf.UpdatedAfter, 10))
	}

	return v
}

// ApplyToQueryBuilder applies the query filter to a query builder.
func (qf *QueryFilter) ApplyToQueryBuilder(queryBuilder squirrel.SelectBuilder, tableName string) squirrel.SelectBuilder {
	if qf == nil {
		return queryBuilder
	}

	const (
		createdOnKey = "created_on"
		updatedOnKey = "last_updated_on"
	)

	qf.SetPage(qf.Page)
	if qp := qf.QueryPage(); qp > 0 {
		queryBuilder = queryBuilder.Offset(qp)
	}

	if qf.Limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(qf.Limit))
	} else {
		queryBuilder = queryBuilder.Limit(MaxLimit)
	}

	if qf.CreatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, createdOnKey): qf.CreatedAfter})
	}

	if qf.CreatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, createdOnKey): qf.CreatedBefore})
	}

	if qf.UpdatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{fmt.Sprintf("%s.%s", tableName, updatedOnKey): qf.UpdatedAfter})
	}

	if qf.UpdatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{fmt.Sprintf("%s.%s", tableName, updatedOnKey): qf.UpdatedBefore})
	}

	return queryBuilder
}

// ExtractQueryFilter can extract a QueryFilter from a request.
func ExtractQueryFilter(req *http.Request) *QueryFilter {
	qf := &QueryFilter{}
	qf.FromParams(req.URL.Query())
	return qf
}
