package postgres

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func buildTestService(t *testing.T) (*Postgres, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	p := ProvidePostgres(true, db, noop.ProvideNoopLogger())
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
	)
	queryArgRegexp = regexp.MustCompile(`\$\d+`)
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
