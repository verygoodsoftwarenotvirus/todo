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
	maxLimit = 50
)

// QueryFilter represents all the filters a user could apply to a list query
type QueryFilter struct {
	Page          uint64   `json:"page"`
	Limit         uint64   `json:"limit"`
	CreatedAfter  uint64   `json:"created_before,omitempty"`
	CreatedBefore uint64   `json:"created_after,omitempty"`
	UpdatedAfter  uint64   `json:"updated_before,omitempty"`
	UpdatedBefore uint64   `json:"updated_after,omitempty"`
	IncludeAll    bool     `json:"include_all,omitempty"`
	SortBy        sortType `json:"sort_by"`
}

// DefaultQueryFilter represents the standard filter collection
var DefaultQueryFilter = buildDefaultQueryFilter()

func buildDefaultQueryFilter() *QueryFilter {
	return &QueryFilter{
		Page:   1,
		Limit:  DefaultLimit,
		SortBy: SortAscending,
	}
}

// FromParams overrides the core QueryFilter values with values retrieved from url.Params
func (qf *QueryFilter) FromParams(params url.Values) {
	if i, err := strconv.ParseUint(params.Get("page"), 10, 64); err == nil {
		qf.Page = uint64(math.Max(float64(i), 1))
	}

	if i, err := strconv.ParseUint(params.Get("limit"), 10, 64); err == nil {
		qf.Limit = uint64(math.Max(float64(i), 1))
	}

	if i, err := strconv.ParseUint(params.Get("created_before"), 10, 64); err == nil {
		qf.CreatedBefore = uint64(i)
	}

	if i, err := strconv.ParseUint(params.Get("created_after"), 10, 64); err == nil {
		qf.CreatedAfter = uint64(i)
	}

	if i, err := strconv.ParseUint(params.Get("updated_before"), 10, 64); err == nil {
		qf.UpdatedAfter = uint64(i)
	}

	if i, err := strconv.ParseUint(params.Get("updated_after"), 10, 64); err == nil {
		qf.UpdatedAfter = uint64(i)
	}

	if ia, err := strconv.ParseBool(params.Get("include_all")); err == nil {
		qf.IncludeAll = ia
	}

	switch strings.ToLower(params.Get("sort_by")) {
	case strSortAsc:
		qf.SortBy = SortAscending
	case strSortDesc:
		qf.SortBy = SortDescending
	}
}

// SetPage sets the current page with certain constraints
func (qf *QueryFilter) SetPage(page uint64) {
	qf.Page = uint64(math.Max(1, float64(page)))
}

// QueryPage calculates a query page from the current filter values
func (qf *QueryFilter) QueryPage() uint {
	return uint(qf.Limit * (qf.Page - 1))
}

// ToValues returns a url.Values from a QueryFilter
func (qf *QueryFilter) ToValues() url.Values {
	if qf == nil {
		return DefaultQueryFilter.ToValues()
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

// ExtractQueryFilter can extract a QueryFilter from a request
func ExtractQueryFilter(req *http.Request) *QueryFilter {
	qf := buildDefaultQueryFilter()
	qf.FromParams(req.URL.Query())
	return qf
}

// BuildQueryLimits does a thing DOCUMENTME
func BuildQueryLimits(pf squirrel.PlaceholderFormat, filter *QueryFilter, creationTimeColumnName, updatedTimeColumnName string) string {
	filter.Limit = uint64(math.Max(float64(filter.Limit), maxLimit))
	filter.Page = uint64(math.Max(1, float64(filter.Page)))
	queryPage := uint64(filter.Limit * (filter.Page - 1))

	sb := squirrel.Select("*")
	if pf != nil {
		sb = sb.PlaceholderFormat(pf)
	}

	if filter.Limit != 0 {
		sb = sb.Limit(filter.Limit)
	}

	if filter.Page != 0 {
		sb = sb.Offset(queryPage)
	}

	if filter.CreatedAfter != 0 {
		sb = sb.Where(squirrel.GtOrEq{creationTimeColumnName: filter.CreatedAfter})
	}

	if filter.CreatedBefore != 0 {
		sb = sb.Where(squirrel.LtOrEq{creationTimeColumnName: filter.CreatedBefore})
	}

	if filter.UpdatedAfter != 0 {
		sb = sb.Where(squirrel.GtOrEq{updatedTimeColumnName: filter.UpdatedAfter})
	}

	if filter.UpdatedBefore != 0 {
		sb = sb.Where(squirrel.LtOrEq{updatedTimeColumnName: filter.UpdatedBefore})
	}

	if filter.SortBy != "" {
		sb = sb.OrderBy(string(filter.SortBy))
	}

	s, _, _ := sb.ToSql()
	return strings.Replace(s, "SELECT *", "", 1)
}
