package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type idMatcher struct{}

func (s *idMatcher) Match(v driver.Value) bool {
	x, ok := v.(string)
	if !ok {
		return false
	}

	if _, err := ksuid.Parse(x); err != nil {
		return false
	}

	return true
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

func newCountDBRowResponse(count uint64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(count)
}

func newSuccessfulDatabaseResult(returnID uint64) driver.Result {
	return sqlmock.NewResult(int64(returnID), 1)
}

func newArbitraryDatabaseResult(_ string) driver.Result {
	return sqlmock.NewResult(1, 1)
}

func formatQueryForSQLMock(query string) string {
	return strings.NewReplacer(
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
	).Replace(query)
}

func interfaceToDriverValue(in []interface{}) []driver.Value {
	out := []driver.Value{}

	for _, x := range in {
		out = append(out, driver.Value(x))
	}

	return out
}

type sqlmockExpecterWrapper struct {
	sqlmock.Sqlmock
}

func (e *sqlmockExpecterWrapper) AssertExpectations(t mock.TestingT) bool {
	return assert.NoError(t, e.Sqlmock.ExpectationsWereMet(), "not all database expectations were met")
}

func buildTestClient(t *testing.T) (*SQLQuerier, *sqlmockExpecterWrapper) {
	t.Helper()

	fakeDB, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	c := &SQLQuerier{
		db:         fakeDB,
		logger:     logging.NewNoopLogger(),
		timeFunc:   defaultTimeFunc,
		tracer:     tracing.NewTracer("test"),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
	}

	return c, &sqlmockExpecterWrapper{Sqlmock: sqlMock}
}

func buildErroneousMockRow() *sqlmock.Rows {
	exampleRows := sqlmock.NewRows([]string{"columns", "don't", "match", "lol"}).AddRow(
		"doesn't",
		"matter",
		"what",
		"goes",
	)

	return exampleRows
}

// end helper funcs

func TestQuerier_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectPing().WillDelayFor(0)

		assert.True(t, c.IsReady(ctx, 1))
	})

	T.Run("with error pinging database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectPing().WillReturnError(errors.New("blah"))

		assert.False(t, c.IsReady(ctx, 1))
	})

	T.Run("exhausting all available queries", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		c, db := buildTestClient(t)

		c.IsReady(ctx, 1)

		db.ExpectPing().WillReturnError(errors.New("blah"))

		assert.False(t, c.IsReady(ctx, 1))
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleConfig := &config.Config{
			Debug:           true,
			RunMigrations:   false,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), exampleConfig, true)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with PostgresProvider", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleConfig := &config.Config{
			Provider:        config.PostgresProvider,
			Debug:           true,
			RunMigrations:   false,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNoopLogger(), exampleConfig, true)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestDefaultTimeFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotZero(t, defaultTimeFunc())
	})
}

func TestQuerier_currentTime(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)

		assert.NotEmpty(t, c.currentTime())
	})

	T.Run("handles nil", func(t *testing.T) {
		t.Parallel()

		var c *SQLQuerier

		assert.NotEmpty(t, c.currentTime())
	})
}

func TestQuerier_rollbackTransaction(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()
		db.ExpectRollback().WillReturnError(errors.New("blah"))

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		c.rollbackTransaction(ctx, tx)
	})
}

func TestQuerier_handleRows(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		mockRows := &database.MockResultIterator{}
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(nil)

		c, _ := buildTestClient(t)

		err := c.checkRowsForErrorAndClose(ctx, mockRows)
		assert.NoError(t, err)
	})

	T.Run("with row error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		expected := errors.New("blah")

		mockRows := &database.MockResultIterator{}
		mockRows.On("Err").Return(expected)

		c, _ := buildTestClient(t)

		err := c.checkRowsForErrorAndClose(ctx, mockRows)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expected))
	})

	T.Run("with close error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		expected := errors.New("blah")

		mockRows := &database.MockResultIterator{}
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(expected)

		c, _ := buildTestClient(t)

		err := c.checkRowsForErrorAndClose(ctx, mockRows)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expected))
	})
}

func TestQuerier_performCreateQueryIgnoringReturn(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(1))

		err := c.performWriteQuery(ctx, c.db, "example", fakeQuery, fakeArgs)

		assert.NoError(t, err)
	})
}

func TestQuerier_performCreateQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(1))

		err := c.performWriteQuery(ctx, c.db, "example", fakeQuery, fakeArgs)

		assert.NoError(t, err)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		err := c.performWriteQuery(ctx, c.db, "example", fakeQuery, fakeArgs)

		assert.Error(t, err)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(int64(1), 0))

		err := c.performWriteQuery(ctx, c.db, "example", fakeQuery, fakeArgs)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})
}
