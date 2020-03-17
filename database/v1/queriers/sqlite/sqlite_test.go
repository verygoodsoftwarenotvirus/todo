package sqlite

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

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
)

func formatQueryForSQLMock(query string) string {
	return sqlMockReplacer.Replace(query)
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
		s, _ := buildTestService(t)
		assert.True(t, s.IsReady(context.Background()))
	})
}

func TestSqlite_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s, _ := buildTestService(t)
		s.logQueryBuildingError(errors.New(""))
	})
}
