package types

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

const (
	// MaxLimit is the maximum value for list queries.
	MaxLimit = 250
	// DefaultLimit represents how many results we return in a response by default.
	DefaultLimit = 20

	// SearchQueryKey is the query param key to find search queries in requests.
	SearchQueryKey = "q"
	// LimitQueryKey is the query param key to specify a limit in a query.
	LimitQueryKey = "limit"

	pageQueryKey            = "page"
	createdBeforeQueryKey   = "createdBefore"
	createdAfterQueryKey    = "createdAfter"
	updatedBeforeQueryKey   = "updatedBefore"
	updatedAfterQueryKey    = "updatedAfter"
	includeArchivedQueryKey = "includeArchived"
	sortByQueryKey          = "sortBy"
)

// QueryFilter represents all the filters a User could apply to a list query.
type QueryFilter struct {
	SortBy          sortType `json:"sortBy"`
	Page            uint64   `json:"page"`
	CreatedAfter    uint64   `json:"createdBefore,omitempty"`
	CreatedBefore   uint64   `json:"createdAfter,omitempty"`
	UpdatedAfter    uint64   `json:"updatedBefore,omitempty"`
	UpdatedBefore   uint64   `json:"updatedAfter,omitempty"`
	Limit           uint8    `json:"limit"`
	IncludeArchived bool     `json:"includeArchived,omitempty"`
}

// DefaultQueryFilter builds the default query filter.
func DefaultQueryFilter() *QueryFilter {
	return &QueryFilter{
		Page:   1,
		Limit:  DefaultLimit,
		SortBy: SortAscending,
	}
}

// AttachToLogger attaches a QueryFilter's values to a logging.Logger.
func (qf *QueryFilter) AttachToLogger(logger logging.Logger) logging.Logger {
	l := logging.EnsureLogger(logger).Clone()

	if qf == nil {
		return l.WithValue(keys.FilterIsNilKey, true)
	}

	if qf.Page != 0 {
		l = l.WithValue(pageQueryKey, qf.Page)
	}

	if qf.Limit != 0 {
		l = l.WithValue(LimitQueryKey, qf.Limit)
	}

	if qf.SortBy != "" {
		l = l.WithValue(sortByQueryKey, qf.SortBy)
	}

	if qf.CreatedBefore != 0 {
		l = l.WithValue(createdBeforeQueryKey, qf.CreatedBefore)
	}

	if qf.CreatedAfter != 0 {
		l = l.WithValue(createdAfterQueryKey, qf.CreatedAfter)
	}

	if qf.UpdatedBefore != 0 {
		l = l.WithValue(updatedBeforeQueryKey, qf.UpdatedBefore)
	}

	if qf.UpdatedAfter != 0 {
		l = l.WithValue(updatedAfterQueryKey, qf.UpdatedAfter)
	}

	return l
}

// FromParams overrides the core QueryFilter values with values retrieved from url.Params.
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

	if i, err := strconv.ParseBool(params.Get(includeArchivedQueryKey)); err == nil {
		qf.IncludeArchived = i
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

// ToValues returns a url.Values from a QueryFilter.
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

	v.Set(includeArchivedQueryKey, strconv.FormatBool(qf.IncludeArchived))

	return v
}

// ExtractQueryFilter can extract a QueryFilter from a request.
func ExtractQueryFilter(req *http.Request) *QueryFilter {
	qf := &QueryFilter{}
	qf.FromParams(req.URL.Query())

	return qf
}
