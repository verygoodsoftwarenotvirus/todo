package postgres

import (
	"context"
	"errors"
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

func TestPostgres_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		p, _ := buildTestService(t)
		assert.True(t, p.IsReady(context.Background()))
	})
}

func Test_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		p, _ := buildTestService(t)
		p.logQueryBuildingError(errors.New(""))
	})
}
