package models

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromParams(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &QueryFilter{
			Page:          100,
			Limit:         50,
			CreatedAfter:  123456789,
			CreatedBefore: 123456789,
			UpdatedAfter:  123456789,
			UpdatedBefore: 123456789,
			SortBy:        SortDescending,
		}

		actual := &QueryFilter{}

		exampleInput := url.Values{
			pageKey: []string{
				strconv.Itoa(int(expected.Page)),
			},
			limitKey: []string{
				strconv.Itoa(int(expected.Limit)),
			},
			createdBeforeKey: []string{
				strconv.Itoa(int(expected.CreatedAfter)),
			},
			createdAfterKey: []string{
				strconv.Itoa(int(expected.CreatedBefore)),
			},
			updatedBeforeKey: []string{
				strconv.Itoa(int(expected.UpdatedAfter)),
			},
			updatedAfterKey: []string{
				strconv.Itoa(int(expected.UpdatedBefore)),
			},
			sortByKey: []string{string(expected.SortBy)},
		}

		actual.FromParams(exampleInput)

		assert.Equal(t, expected, actual)

		exampleInput[sortByKey] = []string{string(SortAscending)}

		actual.FromParams(exampleInput)

		assert.Equal(t, SortAscending, actual.SortBy)
	})
}

func TestQueryFilter_SetPage(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		qf := &QueryFilter{}
		expected := uint64(123)
		qf.SetPage(expected)
		assert.Equal(t, expected, qf.Page)
	})
}

func TestQueryFilter_QueryPage(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		qf := &QueryFilter{
			Limit: 10,
			Page:  11,
		}
		expected := uint64(100)
		actual := qf.QueryPage()
		assert.Equal(t, expected, actual)
	})
}

func TestQueryFilter_ToValues(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {

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
			pageKey: []string{
				strconv.Itoa(int(qf.Page)),
			},
			limitKey: []string{
				strconv.Itoa(int(qf.Limit)),
			},
			createdBeforeKey: []string{
				strconv.Itoa(int(qf.CreatedAfter)),
			},
			createdAfterKey: []string{
				strconv.Itoa(int(qf.CreatedBefore)),
			},
			updatedBeforeKey: []string{
				strconv.Itoa(int(qf.UpdatedAfter)),
			},
			updatedAfterKey: []string{
				strconv.Itoa(int(qf.UpdatedBefore)),
			},
			sortByKey: []string{string(qf.SortBy)},
		}

		actual := qf.ToValues()

		assert.Equal(t, expected, actual)
	})
}
