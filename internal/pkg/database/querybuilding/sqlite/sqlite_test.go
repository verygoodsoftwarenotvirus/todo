package sqlite

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*Sqlite, sqlmock.Sqlmock) {
	_, mock, err := sqlmock.New()
	require.NoError(t, err)

	q := ProvideSqlite(logging.NewNonOperationalLogger())

	return q, mock
}

func assertArgCountMatchesQuery(t *testing.T, query string, args []interface{}) {
	t.Helper()

	queryArgCount := len(regexp.MustCompile(`\?+`).FindAllString(query, -1))

	if len(args) > 0 {
		assert.Equal(t, queryArgCount, len(args))
	} else {
		assert.Zero(t, queryArgCount)
	}
}

func TestProvideSqlite(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		buildTestService(t)
	})
}

func TestSqlite_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()
		_, span := tracing.StartSpan(ctx)

		q.logQueryBuildingError(span, errors.New("blah"))
	})
}

func TestProvideSqliteDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideSqliteDB(logging.NewNonOperationalLogger(), "", time.Hour)
		assert.NoError(t, err)
	})
}
