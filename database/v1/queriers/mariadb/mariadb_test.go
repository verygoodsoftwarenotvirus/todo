package mariadb

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) (*MariaDB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	p := ProvideMariaDB(true, db, noop.ProvideNoopLogger())
	return p.(*MariaDB), mock
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

func TestProvidePostgres(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		buildTestService(t)
	})
}

func TestMariaDB_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		m, _ := buildTestService(t)
		assert.True(t, m.IsReady(context.Background()))
	})

}

func Test_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		m, _ := buildTestService(t)
		m.logQueryBuildingError(errors.New(""))
	})
}
