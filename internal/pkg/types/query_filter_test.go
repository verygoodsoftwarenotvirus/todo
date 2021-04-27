package types

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryFilter_AttachToLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		qf := &QueryFilter{
			Page:            100,
			Limit:           MaxLimit,
			CreatedAfter:    123456789,
			CreatedBefore:   123456789,
			UpdatedAfter:    123456789,
			UpdatedBefore:   123456789,
			SortBy:          SortDescending,
			IncludeArchived: true,
		}

		assert.NotNil(t, qf.AttachToLogger(logger))
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()

		assert.NotNil(t, (*QueryFilter)(nil).AttachToLogger(logger))
	})
}

func TestQueryFilter_FromParams(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		actual := &QueryFilter{}
		expected := &QueryFilter{
			Page:            100,
			Limit:           MaxLimit,
			CreatedAfter:    123456789,
			CreatedBefore:   123456789,
			UpdatedAfter:    123456789,
			UpdatedBefore:   123456789,
			SortBy:          SortDescending,
			IncludeArchived: true,
		}

		exampleInput := url.Values{
			pageQueryKey:            []string{strconv.Itoa(int(expected.Page))},
			LimitQueryKey:           []string{strconv.Itoa(int(expected.Limit))},
			createdBeforeQueryKey:   []string{strconv.Itoa(int(expected.CreatedAfter))},
			createdAfterQueryKey:    []string{strconv.Itoa(int(expected.CreatedBefore))},
			updatedBeforeQueryKey:   []string{strconv.Itoa(int(expected.UpdatedAfter))},
			updatedAfterQueryKey:    []string{strconv.Itoa(int(expected.UpdatedBefore))},
			sortByQueryKey:          []string{string(expected.SortBy)},
			includeArchivedQueryKey: []string{strconv.FormatBool(true)},
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

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		qf := &QueryFilter{}
		expected := uint64(123)
		qf.SetPage(expected)

		assert.Equal(t, expected, qf.Page)
	})
}

func TestQueryFilter_QueryPage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		qf := &QueryFilter{Limit: 10, Page: 11}
		expected := uint64(100)
		actual := qf.QueryPage()

		assert.Equal(t, expected, actual)
	})
}

func TestQueryFilter_ToValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		qf := &QueryFilter{
			Page:            100,
			Limit:           50,
			CreatedAfter:    123456789,
			CreatedBefore:   123456789,
			UpdatedAfter:    123456789,
			UpdatedBefore:   123456789,
			IncludeArchived: true,
			SortBy:          SortDescending,
		}
		expected := url.Values{
			pageQueryKey:            []string{strconv.Itoa(int(qf.Page))},
			LimitQueryKey:           []string{strconv.Itoa(int(qf.Limit))},
			createdBeforeQueryKey:   []string{strconv.Itoa(int(qf.CreatedAfter))},
			createdAfterQueryKey:    []string{strconv.Itoa(int(qf.CreatedBefore))},
			updatedBeforeQueryKey:   []string{strconv.Itoa(int(qf.UpdatedAfter))},
			updatedAfterQueryKey:    []string{strconv.Itoa(int(qf.UpdatedBefore))},
			includeArchivedQueryKey: []string{strconv.FormatBool(qf.IncludeArchived)},
			sortByQueryKey:          []string{string(qf.SortBy)},
		}

		actual := qf.ToValues()
		assert.Equal(t, expected, actual)
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()
		qf := (*QueryFilter)(nil)
		expected := DefaultQueryFilter().ToValues()
		actual := qf.ToValues()
		assert.Equal(t, expected, actual)
	})
}

func TestExtractQueryFilter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://verygoodsoftwarenotvirus.ru", nil)
		assert.NoError(t, err)
		require.NotNil(t, req)

		req.URL.RawQuery = exampleInput.Encode()
		actual := ExtractQueryFilter(req)
		assert.Equal(t, expected, actual)
	})
}
