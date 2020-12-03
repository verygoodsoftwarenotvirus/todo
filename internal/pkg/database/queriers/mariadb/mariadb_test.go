package mariadb

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

func buildTestService(t *testing.T) (*MariaDB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return ProvideMariaDB(true, db, noop.NewLogger()).(*MariaDB), mock
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
	queryArgRegexp = regexp.MustCompile(`\?+`)
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

func assertArgCountMatchesQuery(t *testing.T, query string, args []interface{}) {
	t.Helper()

	queryArgCount := len(queryArgRegexp.FindAllString(query, -1))

	if len(args) > 0 {
		assert.Equal(t, queryArgCount, len(args))
	} else {
		assert.Zero(t, queryArgCount)
	}
}

func TestProvideMariaDB(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		buildTestService(t)
	})
}

func TestMariaDB_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, _ := buildTestService(t)
		assert.True(t, q.IsReady(ctx))
	})
}

func TestMariaDB_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		q.logQueryBuildingError(errors.New("blah"))
	})
}

func TestMariaDB_logIDRetrievalError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		q.logIDRetrievalError(errors.New("blah"))
	})
}

func TestProvideMariaDBConnection(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideMariaDBConnection(noop.NewLogger(), "")
		assert.NoError(t, err)
	})
}
