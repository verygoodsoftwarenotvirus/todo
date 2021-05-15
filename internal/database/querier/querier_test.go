package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	defaultLimit = uint8(20)
)

// begin helper funcs

/*
type testHelper struct {
	ctx            context.Context
	filter         *types.QueryFilter
	querier        *SQLQuerier
	mockDB         *sqlmockExpecterWrapper
	exampleUser    *types.User
	exampleAccount *types.Account
}

func buildTestHelper(t *testing.T) *testHelper {
	t.Helper()

	q, db := buildTestClient(t)

	h := &testHelper{
		querier:        q,
		mockDB:         db,
		ctx:            context.Background(),
		exampleUser:    fakes.BuildFakeUser(),
		exampleAccount: fakes.BuildFakeAccount(),
		filter:         types.DefaultQueryFilter(),
	}

	h.exampleAccount.BelongsToUser = h.exampleUser.ID

	return h
}

func (h *testHelper) AssertExpectations(t mock.TestingT) bool {
	return assert.NoError(t, h.mockDB.ExpectationsWereMet(), "not all database expectations were met")
}
*/

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

func buildTestClient(t *testing.T) (*SQLQuerier, *sqlmockExpecterWrapper) {
	t.Helper()

	db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	c := &SQLQuerier{
		db:              db,
		logger:          logging.NewNonOperationalLogger(),
		timeFunc:        defaultTimeFunc,
		tracer:          tracing.NewTracer("test"),
		sqlQueryBuilder: database.BuildMockSQLQueryBuilder(),
		idStrategy:      DefaultIDRetrievalStrategy,
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

func expectAuditLogEntryInTransaction(mockQueryBuilder *database.MockSQLQueryBuilder, db sqlmock.Sqlmock, returnErr error) {
	fakeAuditLogEntryQuery, fakeAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
	mockQueryBuilder.AuditLogEntrySQLQueryBuilder.
		On("BuildCreateAuditLogEntryQuery",
			testutil.ContextMatcher,
			mock.IsType(&types.AuditLogEntryCreationInput{})).
		Return(fakeAuditLogEntryQuery, fakeAuditLogEntryArgs)

	e := db.ExpectExec(formatQueryForSQLMock(fakeAuditLogEntryQuery)).
		WithArgs(interfaceToDriverValue(fakeAuditLogEntryArgs)...)

	if returnErr != nil {
		e.WillReturnError(returnErr)
	} else {
		e.WillReturnResult(newSuccessfulDatabaseResult(123))
	}
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
		var migrationFunctionCalled bool
		fakeMigrationFunc := func() {
			migrationFunctionCalled = true
		}

		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		queryBuilder := database.BuildMockSQLQueryBuilder()
		queryBuilder.On(
			"BuildMigrationFunc",
			mock.IsType(&sql.DB{}),
		).Return(fakeMigrationFunc)

		mockDB.ExpectPing().WillDelayFor(0)

		exampleConfig := &dbconfig.Config{
			Debug:           true,
			RunMigrations:   true,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNonOperationalLogger(), db, exampleConfig, queryBuilder, true)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.True(t, migrationFunctionCalled)

		mock.AssertExpectationsForObjects(t, &sqlmockExpecterWrapper{Sqlmock: mockDB}, queryBuilder)
	})

	T.Run("with PostgresProvider", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		var migrationFunctionCalled bool
		fakeMigrationFunc := func() {
			migrationFunctionCalled = true
		}

		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		queryBuilder := database.BuildMockSQLQueryBuilder()
		queryBuilder.On(
			"BuildMigrationFunc",
			mock.IsType(&sql.DB{}),
		).Return(fakeMigrationFunc)

		mockDB.ExpectPing().WillDelayFor(0)

		exampleConfig := &dbconfig.Config{
			Provider:        dbconfig.PostgresProvider,
			Debug:           true,
			RunMigrations:   true,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNonOperationalLogger(), db, exampleConfig, queryBuilder, true)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.True(t, migrationFunctionCalled)

		mock.AssertExpectationsForObjects(t, &sqlmockExpecterWrapper{Sqlmock: mockDB}, queryBuilder)
	})

	T.Run("with error initializing querier", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		queryBuilder := database.BuildMockSQLQueryBuilder()

		mockDB.ExpectPing().WillReturnError(errors.New("blah"))

		exampleConfig := &dbconfig.Config{
			Debug:           true,
			RunMigrations:   true,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNonOperationalLogger(), db, exampleConfig, queryBuilder, true)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, &sqlmockExpecterWrapper{Sqlmock: mockDB}, queryBuilder)
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

func TestQuerier_getIDFromResult(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := int64(123)

		m := &database.MockSQLResult{}
		m.On("LastInsertId").Return(expected, nil)

		ctx := context.Background()
		c, _ := buildTestClient(t)
		actual := c.getIDFromResult(ctx, m)

		assert.Equal(t, uint64(expected), actual)
	})

	T.Run("logs error", func(t *testing.T) {
		t.Parallel()

		m := &database.MockSQLResult{}
		m.On("LastInsertId").Return(int64(0), errors.New("blah"))

		ctx := context.Background()
		c, _ := buildTestClient(t)
		actual := c.getIDFromResult(ctx, m)

		assert.Zero(t, actual)
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

		err := c.performWriteQueryIgnoringReturn(ctx, c.db, "example", fakeQuery, fakeArgs)

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

		_, err := c.performWriteQuery(ctx, c.db, false, "example", fakeQuery, fakeArgs)

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

		_, err := c.performWriteQuery(ctx, c.db, false, "example", fakeQuery, fakeArgs)

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

		_, err := c.performWriteQuery(ctx, c.db, false, "example", fakeQuery, fakeArgs)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})

	T.Run("with ReturningStatementIDRetrievalStrategy", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)
		c.idStrategy = ReturningStatementIDRetrievalStrategy

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(123)))

		_, err := c.performWriteQuery(ctx, c.db, false, "example", fakeQuery, fakeArgs)

		assert.NoError(t, err)
	})

	T.Run("with ReturningStatementIDRetrievalStrategy and error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)
		c.idStrategy = ReturningStatementIDRetrievalStrategy

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		id, err := c.performWriteQuery(ctx, c.db, false, "example", fakeQuery, fakeArgs)

		assert.Zero(t, id)
		assert.Error(t, err)
	})

	T.Run("ignoring return with return statement id strategy", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)
		c.idStrategy = ReturningStatementIDRetrievalStrategy

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(1))

		_, err := c.performWriteQuery(ctx, c.db, true, "example", fakeQuery, fakeArgs)

		assert.NoError(t, err)
	})

	T.Run("ignoring return with return statement id strategy and error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)
		c.idStrategy = ReturningStatementIDRetrievalStrategy

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		_, err := c.performWriteQuery(ctx, c.db, true, "example", fakeQuery, fakeArgs)

		assert.Error(t, err)
	})

	T.Run("ignoring return with return statement id strategy with no rows affected", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)
		c.idStrategy = ReturningStatementIDRetrievalStrategy

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		_, err := c.performWriteQuery(ctx, c.db, true, "example", fakeQuery, fakeArgs)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})
}