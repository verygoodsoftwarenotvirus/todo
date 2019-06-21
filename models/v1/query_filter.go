package models

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
)

const (
	// MaxLimit is the maximum value for list queries
	MaxLimit = 250
	// DefaultLimit represents how many results we return in a response by default.
	DefaultLimit = 20

	pageKey          = "page"
	limitKey         = "limit"
	createdBeforeKey = "created_before"
	createdAfterKey  = "created_after"
	updatedBeforeKey = "updated_before"
	updatedAfterKey  = "updated_after"
	sortByKey        = "sort_by"
)

// QueryFilter represents all the filters a user could apply to a list query
type QueryFilter struct {
	Page          uint64   `json:"page"`
	Limit         uint64   `json:"limit"`
	CreatedAfter  uint64   `json:"created_before,omitempty"`
	CreatedBefore uint64   `json:"created_after,omitempty"`
	UpdatedAfter  uint64   `json:"updated_before,omitempty"`
	UpdatedBefore uint64   `json:"updated_after,omitempty"`
	SortBy        sortType `json:"sort_by"`
}

// DefaultQueryFilter builds the default query filter
func DefaultQueryFilter() *QueryFilter {
	return &QueryFilter{
		Page:   1,
		Limit:  DefaultLimit,
		SortBy: SortAscending,
	}
}

// FromParams overrides the core QueryFilter values with values retrieved from url.Params
func (qf *QueryFilter) FromParams(params url.Values) {
	if i, err := strconv.ParseUint(params.Get(pageKey), 10, 64); err == nil {
		qf.Page = uint64(math.Max(float64(i), 1))
	}

	if i, err := strconv.ParseUint(params.Get(limitKey), 10, 64); err == nil {
		qf.Limit = uint64(math.Max(math.Max(float64(i), 0), MaxLimit))
	}

	if i, err := strconv.ParseUint(params.Get(createdBeforeKey), 10, 64); err == nil {
		qf.CreatedBefore = uint64(math.Max(float64(i), 0))
	}

	if i, err := strconv.ParseUint(params.Get(createdAfterKey), 10, 64); err == nil {
		qf.CreatedAfter = uint64(math.Max(float64(i), 0))
	}

	if i, err := strconv.ParseUint(params.Get(updatedBeforeKey), 10, 64); err == nil {
		qf.UpdatedBefore = uint64(math.Max(float64(i), 0))
	}

	if i, err := strconv.ParseUint(params.Get(updatedAfterKey), 10, 64); err == nil {
		qf.UpdatedAfter = uint64(math.Max(float64(i), 0))
	}

	switch strings.ToLower(params.Get(sortByKey)) {
	case string(SortAscending):
		qf.SortBy = SortAscending
	case string(SortDescending):
		qf.SortBy = SortDescending
	}
}

// SetPage sets the current page with certain constraints
func (qf *QueryFilter) SetPage(page uint64) {
	qf.Page = uint64(math.Max(1, float64(page)))
}

// QueryPage calculates a query page from the current filter values
func (qf *QueryFilter) QueryPage() uint64 {
	return qf.Limit * (qf.Page - 1)
}

// ToValues returns a url.Values from a QueryFilter
func (qf *QueryFilter) ToValues() url.Values {
	if qf == nil {
		return DefaultQueryFilter().ToValues()
	}

	v := url.Values{}
	if qf.Page != 0 {
		v.Set("page", strconv.FormatUint(qf.Page, 10))
	}
	if qf.Limit != 0 {
		v.Set("limit", strconv.FormatUint(qf.Limit, 10))
	}
	if qf.SortBy != "" {
		v.Set("sort_by", string(qf.SortBy))
	}
	if qf.CreatedBefore != 0 {
		v.Set("created_before", strconv.FormatUint(qf.CreatedBefore, 10))
	}
	if qf.CreatedAfter != 0 {
		v.Set("created_after", strconv.FormatUint(qf.CreatedAfter, 10))
	}
	if qf.UpdatedBefore != 0 {
		v.Set("updated_before", strconv.FormatUint(qf.UpdatedBefore, 10))
	}
	if qf.UpdatedAfter != 0 {
		v.Set("updated_after", strconv.FormatUint(qf.UpdatedAfter, 10))
	}

	return v
}

// ApplyToQueryBuilder applies the query filter to a query builder
func (qf *QueryFilter) ApplyToQueryBuilder(queryBuilder squirrel.SelectBuilder) squirrel.SelectBuilder {
	if qf == nil {
		return queryBuilder
	}
	qf.SetPage(qf.Page)

	if qp := qf.QueryPage(); qp > 0 {
		queryBuilder = queryBuilder.Offset(qp)
	}

	if qf.Limit > 0 {
		queryBuilder = queryBuilder.Limit(qf.Limit)
	} else {
		queryBuilder = queryBuilder.Limit(MaxLimit)
	}

	if qf.CreatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{"created_on": qf.CreatedAfter})
	}

	if qf.CreatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{"created_on": qf.CreatedBefore})
	}

	if qf.UpdatedAfter > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Gt{"updated_on": qf.UpdatedAfter})
	}

	if qf.UpdatedBefore > 0 {
		queryBuilder = queryBuilder.Where(squirrel.Lt{"updated_on": qf.UpdatedBefore})
	}

	return queryBuilder
}

// ExtractQueryFilter can extract a QueryFilter from a request
func ExtractQueryFilter(req *http.Request) *QueryFilter {
	qf := &QueryFilter{}
	qf.FromParams(req.URL.Query())
	return qf
}
