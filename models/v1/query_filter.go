package models

import (
	"errors"
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

type QueryFilter struct {
	Page          uint64   `json:"page"`
	Limit         uint64   `json:"limit"`
	CreatedAfter  uint64   `json:"created_before"`
	CreatedBefore uint64   `json:"created_after"`
	UpdatedAfter  uint64   `json:"updated_before"`
	UpdatedBefore uint64   `json:"updated_after"`
	IncludeAll    bool     `json:"include_all"`
	SortBy        sortType `json:"sort_by"`
}

var DefaultQueryFilter = buildDefaultQueryFilter()

func buildDefaultQueryFilter() *QueryFilter {
	return &QueryFilter{
		Page:   1,
		Limit:  DefaultLimit,
		SortBy: SortAscending,
	}
}

func (qf *QueryFilter) FromParams(params url.Values) {
	if page, ok := params["page"]; ok && len(page) >= 1 {
		if i, err := strconv.ParseUint(page[0], 10, 64); err == nil {
			qf.Page = uint64(math.Max(float64(i), 1))
		}
	}

	if limit, ok := params["limit"]; ok && len(limit) >= 1 {
		if i, err := strconv.ParseUint(limit[0], 10, 64); err == nil {
			qf.Limit = uint64(math.Max(float64(i), 1))
		}
	}

	if createdBefore, ok := params["created_before"]; ok && len(createdBefore) >= 1 {
		if i, err := strconv.ParseUint(createdBefore[0], 10, 64); err == nil {
			qf.CreatedBefore = uint64(i)
		}
	}

	if createdAfter, ok := params["created_after"]; ok && len(createdAfter) >= 1 {
		if i, err := strconv.ParseUint(createdAfter[0], 10, 64); err == nil {
			qf.CreatedAfter = uint64(i)
		}
	}

	if updatedBefore, ok := params["updated_before"]; ok && len(updatedBefore) >= 1 {
		if i, err := strconv.ParseUint(updatedBefore[0], 10, 64); err == nil {
			qf.UpdatedAfter = uint64(i)
		}
	}

	if updatedAfter, ok := params["updated_after"]; ok && len(updatedAfter) >= 1 {
		if i, err := strconv.ParseUint(updatedAfter[0], 10, 64); err == nil {
			qf.UpdatedAfter = uint64(i)
		}
	}

	if includeAll, ok := params["include_all"]; ok && len(includeAll) >= 1 {
		if ia, err := strconv.ParseBool(includeAll[0]); err == nil {
			qf.IncludeAll = ia
		}
	}

	if sortBy, ok := params["sort_by"]; ok && len(sortBy) >= 1 {
		switch strings.ToLower(sortBy[0]) {
		case strSortAsc:
			qf.SortBy = SortAscending
		case strSortDesc:
			qf.SortBy = SortDescending
		}
	}
}

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

func (qf *QueryFilter) ToMap() map[string]string {
	if qf == nil {
		return DefaultQueryFilter.ToMap()
	}

	return map[string]string{
		"page":  strconv.Itoa(int(qf.Page)),
		"limit": strconv.Itoa(int(qf.Limit)),
	}
}

func ParseQueryFilter(req *http.Request) *QueryFilter {
	qf := buildDefaultQueryFilter()
	qf.FromParams(req.URL.Query())
	return qf
}

type QueryLimiter struct {
	CreationTimeColumnName string
	UpdatedTimeColumnName  string
}

func NewQueryLimiter(CreationTimeColumnName string, UpdatedTimeColumnName string) (*QueryLimiter, error) {
	if CreationTimeColumnName == "" || UpdatedTimeColumnName == "" {
		return nil, errors.New("column names must not be nil")
	}
	return &QueryLimiter{
		CreationTimeColumnName: CreationTimeColumnName,
		UpdatedTimeColumnName:  UpdatedTimeColumnName,
	}, nil
}

func (l *QueryLimiter) BuildQueryLimits(pf squirrel.PlaceholderFormat, filter *QueryFilter) string {
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
		sb = sb.Where(squirrel.GtOrEq{l.CreationTimeColumnName: filter.CreatedAfter})
	}

	if filter.CreatedBefore != 0 {
		sb = sb.Where(squirrel.LtOrEq{l.CreationTimeColumnName: filter.CreatedBefore})
	}

	if filter.UpdatedAfter != 0 {
		sb = sb.Where(squirrel.GtOrEq{l.UpdatedTimeColumnName: filter.UpdatedAfter})
	}

	if filter.UpdatedBefore != 0 {
		sb = sb.Where(squirrel.LtOrEq{l.UpdatedTimeColumnName: filter.UpdatedBefore})
	}

	if filter.SortBy != "" {
		sb = sb.OrderBy(string(filter.SortBy))
	}

	s, _, _ := sb.ToSql()
	return strings.Replace(s, "SELECT *", "", 1)
}
