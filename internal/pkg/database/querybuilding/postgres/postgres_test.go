package postgres

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*Postgres, sqlmock.Sqlmock) {
	t.Helper()

	_, mock, err := sqlmock.New()
	require.NoError(t, err)

	return ProvidePostgres(logging.NewNonOperationalLogger()), mock
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
		q.logQueryBuildingError(errors.New("blah"))
	})
}

func Test_joinUint64s(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		exampleInput := []uint64{123, 456, 789}
		expected := "123,456,789"
		actual := joinUint64s(exampleInput)

		assert.Equal(t, expected, actual, "expected %s to equal %s", expected, actual)
	})
}

func TestProvidePostgresDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		_, err := ProvidePostgresDB(logging.NewNonOperationalLogger(), "")
		assert.NoError(t, err)
	})
}
