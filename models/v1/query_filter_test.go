package models

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromParams(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		actual := &QueryFilter{}
		expected := &QueryFilter{
			Page:          100,
			Limit:         MaxLimit,
			CreatedAfter:  123456789,
			CreatedBefore: 123456789,
			UpdatedAfter:  123456789,
			UpdatedBefore: 123456789,
			SortBy:        SortDescending,
		}

		exampleInput := url.Values{
			pageQueryKey:          []string{strconv.Itoa(int(expected.Page))},
			LimitQueryKey:         []string{strconv.Itoa(int(expected.Limit))},
			createdBeforeQueryKey: []string{strconv.Itoa(int(expected.CreatedAfter))},
			createdAfterQueryKey:  []string{strconv.Itoa(int(expected.CreatedBefore))},
			updatedBeforeQueryKey: []string{strconv.Itoa(int(expected.UpdatedAfter))},
			updatedAfterQueryKey:  []string{strconv.Itoa(int(expected.UpdatedBefore))},
			sortByQueryKey:        []string{string(expected.SortBy)},
		}

		actual.FromParams(exampleInput)
		assert.Equal(t, expected, actual)

		exampleInput[sortByQueryKey] = []string{string(SortAscending)}

		actual.FromParams(exampleInput)
		assert.Equal(t, SortAscending, actual.SortBy)
	})
}

func TestQueryFilter_SetPage(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		qf := &QueryFilter{}
		expected := uint64(123)
		qf.SetPage(expected)
		assert.Equal(t, expected, qf.Page)
	})
}

func TestQueryFilter_QueryPage(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		qf := &QueryFilter{Limit: 10, Page: 11}
		expected := uint64(100)
		actual := qf.QueryPage()
		assert.Equal(t, expected, actual)
	})
}

func TestQueryFilter_ToValues(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		qf := &QueryFilter{
			Page:          100,
			Limit:         50,
			CreatedAfter:  123456789,
			CreatedBefore: 123456789,
			UpdatedAfter:  123456789,
			UpdatedBefore: 123456789,
			SortBy:        SortDescending,
		}
		expected := url.Values{
			pageQueryKey:          []string{strconv.Itoa(int(qf.Page))},
			LimitQueryKey:         []string{strconv.Itoa(int(qf.Limit))},
			createdBeforeQueryKey: []string{strconv.Itoa(int(qf.CreatedAfter))},
			createdAfterQueryKey:  []string{strconv.Itoa(int(qf.CreatedBefore))},
			updatedBeforeQueryKey: []string{strconv.Itoa(int(qf.UpdatedAfter))},
			updatedAfterQueryKey:  []string{strconv.Itoa(int(qf.UpdatedBefore))},
			sortByQueryKey:        []string{string(qf.SortBy)},
		}

		actual := qf.ToValues()
		assert.Equal(t, expected, actual)
	})

	T.Run("with nil", func(t *testing.T) {
		qf := (*QueryFilter)(nil)
		expected := DefaultQueryFilter().ToValues()
		actual := qf.ToValues()
		assert.Equal(t, expected, actual)
	})
}

func TestQueryFilter_ApplyToQueryBuilder(T *testing.T) {
	T.Parallel()

	exampleTableName := "stuff"
	baseQueryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("things").
		From(exampleTableName).
		Where(squirrel.Eq{fmt.Sprintf("%s.condition", exampleTableName): true})

	T.Run("happy path", func(t *testing.T) {
		qf := &QueryFilter{
			Page:          100,
			Limit:         50,
			CreatedAfter:  123456789,
			CreatedBefore: 123456789,
			UpdatedAfter:  123456789,
			UpdatedBefore: 123456789,
			SortBy:        SortDescending,
		}

		sb := squirrel.StatementBuilder.Select("*").From("testing")
		qf.ApplyToQueryBuilder(sb, exampleTableName)
		expected := "SELECT * FROM testing"
		actual, _, err := sb.ToSql()

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("basic usecase", func(t *testing.T) {
		exampleQF := &QueryFilter{Limit: 15, Page: 2}

		expected := "SELECT things FROM stuff WHERE stuff.condition = $1 LIMIT 15 OFFSET 15"
		x := exampleQF.ApplyToQueryBuilder(baseQueryBuilder, exampleTableName)
		actual, args, err := x.ToSql()

		assert.Equal(t, expected, actual, "expected and actual queries don't match")
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
	})

	T.Run("returns query builder if query filter is nil", func(t *testing.T) {
		expected := "SELECT things FROM stuff WHERE stuff.condition = $1"

		x := (*QueryFilter)(nil).ApplyToQueryBuilder(baseQueryBuilder, exampleTableName)
		actual, args, err := x.ToSql()

		assert.Equal(t, expected, actual, "expected and actual queries don't match")
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
	})

	T.Run("whole kit and kaboodle", func(t *testing.T) {
		exampleQF := &QueryFilter{
			Limit:         20,
			Page:          6,
			CreatedAfter:  uint64(time.Now().Unix()),
			CreatedBefore: uint64(time.Now().Unix()),
			UpdatedAfter:  uint64(time.Now().Unix()),
			UpdatedBefore: uint64(time.Now().Unix()),
		}

		expected := "SELECT things FROM stuff WHERE stuff.condition = $1 AND stuff.created_on > $2 AND stuff.created_on < $3 AND stuff.last_updated_on > $4 AND stuff.last_updated_on < $5 LIMIT 20 OFFSET 100"
		x := exampleQF.ApplyToQueryBuilder(baseQueryBuilder, exampleTableName)
		actual, args, err := x.ToSql()

		assert.Equal(t, expected, actual, "expected and actual queries don't match")
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
	})

	T.Run("with zero limit", func(t *testing.T) {
		exampleQF := &QueryFilter{Limit: 0, Page: 1}
		expected := "SELECT things FROM stuff WHERE stuff.condition = $1 LIMIT 250"
		x := exampleQF.ApplyToQueryBuilder(baseQueryBuilder, exampleTableName)
		actual, args, err := x.ToSql()

		assert.Equal(t, expected, actual, "expected and actual queries don't match")
		assert.Nil(t, err)
		assert.NotEmpty(t, args)
	})
}

func TestExtractQueryFilter(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &QueryFilter{
			Page:          100,
			Limit:         MaxLimit,
			CreatedAfter:  123456789,
			CreatedBefore: 123456789,
			UpdatedAfter:  123456789,
			UpdatedBefore: 123456789,
			SortBy:        SortDescending,
		}
		exampleInput := url.Values{
			pageQueryKey:          []string{strconv.Itoa(int(expected.Page))},
			LimitQueryKey:         []string{strconv.Itoa(int(expected.Limit))},
			createdBeforeQueryKey: []string{strconv.Itoa(int(expected.CreatedAfter))},
			createdAfterQueryKey:  []string{strconv.Itoa(int(expected.CreatedBefore))},
			updatedBeforeQueryKey: []string{strconv.Itoa(int(expected.UpdatedAfter))},
			updatedAfterQueryKey:  []string{strconv.Itoa(int(expected.UpdatedBefore))},
			sortByQueryKey:        []string{string(expected.SortBy)},
		}

		req, err := http.NewRequest(http.MethodGet, "https://verygoodsoftwarenotvirus.ru", nil)
		assert.NoError(t, err)
		require.NotNil(t, req)

		req.URL.RawQuery = exampleInput.Encode()
		actual := ExtractQueryFilter(req)
		assert.Equal(t, expected, actual)
	})
}
