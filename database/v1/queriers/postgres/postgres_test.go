package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"
)

func buildTestService(t *testing.T) (*Postgres, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	p := ProvidePostgres(true, db, noop.ProvideNoopLogger(), "")
	return p.(*Postgres), mock
}

func formatQueryForSQLMock(query string) string {
	for _, x := range []string{"$", "(", ")", "=", "*", ".", "+", "?", ",", "-"} {
		query = strings.Replace(query, x, fmt.Sprintf(`\%s`, x), -1)
	}
	return query
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
		logQueryBuildingError(noop.ProvideNoopLogger(), errors.New(""))
	})
}
