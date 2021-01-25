package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

const (
	defaultLimit = uint8(20)
)

// begin helper funcs

func newCountDBRowResponse(count uint64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(count)
}

func newSuccessfulDatabaseResult(returnID uint64) driver.Result {
	return sqlmock.NewResult(int64(returnID), 1)
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

func buildTestClient(t *testing.T) (*Client, *sqlmockExpecterWrapper) {
	t.Helper()

	db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	c := &Client{
		db:              db,
		logger:          noop.NewLogger(),
		timeFunc:        defaultTimeFunc,
		tracer:          tracing.NewTracer("test"),
		sqlQueryBuilder: database.BuildMockSQLQueryBuilder(),
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

func TestClient_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		c.Migrate(ctx, nil)
	})
}

func TestClient_IsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

		select {
		case <-time.After(1 * time.Second):
			assert.Fail(t, "expired")
		case <-ctx.Done():
			assert.True(t, true)
		}

		db.ExpectPing().WillDelayFor(1034 * time.Minute)

		assert.False(t, c.IsReady(ctx, 1))
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	exampleConfig := &dbconfig.Config{
		Debug:         true,
		RunMigrations: true,
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(nil)

		actual, err := ProvideDatabaseClient(ctx, noop.NewLogger(), mockDB, nil, exampleConfig)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error migrating querier", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		expected := errors.New("blah")
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(expected)

		x, actual := ProvideDatabaseClient(ctx, noop.NewLogger(), mockDB, nil, exampleConfig)
		assert.Nil(t, x)
		assert.Error(t, actual)
		assert.Equal(t, expected, errors.Unwrap(actual))

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestDefaultTimeFunc(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		assert.NotZero(t, defaultTimeFunc())
	})
}

func TestClient_currentTime(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)

		assert.NotEmpty(t, c.currentTime())
	})

	T.Run("hadnles nil", func(t *testing.T) {
		t.Parallel()

		var c *Client

		assert.NotEmpty(t, c.currentTime())
	})
}

func TestClient_getIDFromResult(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		expected := int64(123)

		m := &database.MockSQLResult{}
		m.On("LastInsertId").Return(expected, nil)

		c, _ := buildTestClient(t)
		actual := c.getIDFromResult(m)

		assert.Equal(t, uint64(expected), actual)
	})

	T.Run("logs error", func(t *testing.T) {
		t.Parallel()

		m := &database.MockSQLResult{}
		m.On("LastInsertId").Return(int64(0), errors.New("blah"))

		c, _ := buildTestClient(t)
		actual := c.getIDFromResult(m)

		assert.Zero(t, actual)
	})
}

func TestClient_handleRows(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		mockRows := &database.MockResultIterator{}
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(nil)

		c, _ := buildTestClient(t)

		err := c.handleRows(mockRows)
		assert.NoError(t, err)
	})

	T.Run("with row error", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("blah")

		mockRows := &database.MockResultIterator{}
		mockRows.On("Err").Return(expected)

		c, _ := buildTestClient(t)

		err := c.handleRows(mockRows)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expected))
	})

	T.Run("with close error", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("blah")

		mockRows := &database.MockResultIterator{}
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(expected)

		c, _ := buildTestClient(t)

		err := c.handleRows(mockRows)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expected))
	})
}

func TestClient_execContext(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(1))

		err := c.execContext(ctx, "example", fakeQuery, fakeArgs)

		assert.NoError(t, err)
	})
}

func TestClient_execContextAndReturnResult(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(1))

		_, err := c.execContextAndReturnResult(ctx, "example", fakeQuery, fakeArgs)

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

		_, err := c.execContextAndReturnResult(ctx, "example", fakeQuery, fakeArgs)

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

		_, err := c.execContextAndReturnResult(ctx, "example", fakeQuery, fakeArgs)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})
}
