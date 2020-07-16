package sqlite

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*Sqlite, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	s := ProvideSqlite(true, db, noop.ProvideNoopLogger())
	return s.(*Sqlite), mock
}

var (
	sqlMockReplacer = strings.NewReplacer(
		"$", `\$`,
		"(", `\(`,
		")", `\)`,
		"=", `\=`,
		"*", `\*`,
		".", `\.`,
		"+", `\+`,
		"?", `\?`,
		",", `\,`,
		"-", `\-`,
	)
	queryArgRegexp = regexp.MustCompile(`\?+`)
)

func formatQueryForSQLMock(query string) string {
	return sqlMockReplacer.Replace(query)
}

func ensureArgCountMatchesQuery(t *testing.T, query string, args []interface{}) {
	t.Helper()

	queryArgCount := len(queryArgRegexp.FindAllString(query, -1))

	if len(args) > 0 {
		assert.Equal(t, queryArgCount, len(args))
	} else {
		assert.Zero(t, queryArgCount)
	}
}

func interfacesToDriverValues(in []interface{}) (out []driver.Value) {
	for _, x := range in {
		out = append(out, driver.Value(x))
	}
	return out
}

func TestProvideSqlite(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		buildTestService(t)
	})
}

func TestSqlite_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		s, _ := buildTestService(t)
		assert.True(t, s.IsReady(ctx))
	})
}

func TestSqlite_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s, _ := buildTestService(t)
		s.logQueryBuildingError(errors.New("blah"))
	})
}

func TestSqlite_logIDRetrievalError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s, _ := buildTestService(t)
		s.logIDRetrievalError(errors.New("blah"))
	})
}

func Test_joinUint64s(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := []uint64{123, 456, 789}

		expected := "123,456,789"
		actual := joinUint64s(exampleInput, ",")

		assert.Equal(t, expected, actual, "expected %s to equal %s", expected, actual)
	})
}

func TestProvideSqliteDB(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_, err := ProvideSqliteDB(noop.ProvideNoopLogger(), "")
		assert.NoError(t, err)
	})
}
