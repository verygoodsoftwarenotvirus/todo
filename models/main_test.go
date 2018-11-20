package models_test

import (
	"strconv"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/stretchr/testify/assert"
)

func TestQueryFilter_ToMap(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		example := &models.QueryFilter{
			Page: 123, Limit: 321,
		}

		expected := map[string]string{
			"page":  strconv.Itoa(int(example.Page)),
			"limit": strconv.Itoa(int(example.Limit)),
		}
		actual := example.ToMap()

		assert.Equal(t, expected, actual)
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()

		example := (*models.QueryFilter)(nil)
		actual := example.ToMap()

		assert.Equal(t, models.DefaultQueryFilter.ToMap(), actual)
	})
}
