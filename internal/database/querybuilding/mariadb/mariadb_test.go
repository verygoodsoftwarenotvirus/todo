package mariadb

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*MariaDB, sqlmock.Sqlmock) {
	t.Helper()

	_, mock, err := sqlmock.New()
	require.NoError(t, err)

	return ProvideMariaDB(logging.NewNoopLogger()), mock
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

func TestProvideMariaDB(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		buildTestService(t)
	})
}

func TestMariaDB_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()
		_, span := tracing.StartSpan(ctx)

		q.logQueryBuildingError(span, errors.New("blah"))
	})
}

func TestProvideMariaDBConnection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideMariaDBConnection(logging.NewNoopLogger(), "")
		assert.NoError(t, err)
	})
}
