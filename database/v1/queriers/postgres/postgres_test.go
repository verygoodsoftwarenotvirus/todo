package postgres

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
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*Postgres, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	p := ProvidePostgres(true, db, noop.NewLogger())
	return p.(*Postgres), mock
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
		"[", `\[`,
		"]", `\]`,
	)
	queryArgRegexp = regexp.MustCompile(`\$\d+`)
)

func interfaceToDriverValue(in []interface{}) []driver.Value {
	out := []driver.Value{}

	for _, x := range in {
		out = append(out, driver.Value(x))
	}

	return out
}

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

func TestProvidePostgres(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		buildTestService(t)
	})
}

func TestPostgres_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		p, _ := buildTestService(t)
		assert.True(t, p.IsReady(ctx))
	})
}

func TestPostgres_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		p, _ := buildTestService(t)
		p.logQueryBuildingError(errors.New("blah"))
	})
}

func Test_joinUint64s(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := []uint64{123, 456, 789}

		expected := "123,456,789"
		actual := joinUint64s(exampleInput)

		assert.Equal(t, expected, actual, "expected %s to equal %s", expected, actual)
	})
}

func TestProvidePostgresDB(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		_, err := ProvidePostgresDB(noop.NewLogger(), "")
		assert.NoError(t, err)
	})
}
