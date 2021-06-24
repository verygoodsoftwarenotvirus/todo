package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*Postgres, sqlmock.Sqlmock) {
	t.Helper()

	_, mock, err := sqlmock.New()
	require.NoError(t, err)

	return ProvidePostgres(logging.NewNoopLogger()), mock
}

func assertArgCountMatchesQuery(t *testing.T, query string, args []interface{}) {
	t.Helper()

	queryArgCount := len(regexp.MustCompile(`\$\d+`).FindAllString(query, -1))

	if len(args) > 0 {
		assert.Equal(t, queryArgCount, len(args))
	} else {
		assert.Zero(t, queryArgCount)
	}
}

func TestProvidePostgres(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		buildTestService(t)
	})
}

func TestPostgres_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()
		_, span := tracing.StartSpan(ctx)

		q.logQueryBuildingError(span, errors.New("blah"))
	})
}

func Test_joinUint64s(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := []uint64{123, 456, 789}
		expected := "123,456,789"
		actual := joinIDsForQuery(exampleInput)

		assert.Equal(t, expected, actual, "expected %s to equal %s", expected, actual)
	})
}

func TestProvidePostgresDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, err := ProvidePostgresDB(logging.NewNoopLogger(), "")
		assert.NoError(t, err)
	})
}
