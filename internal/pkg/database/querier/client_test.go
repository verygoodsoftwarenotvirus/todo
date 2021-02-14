package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func expectAuditLogEntryInTransaction(mockQueryBuilder *database.MockSQLQueryBuilder, db sqlmock.Sqlmock) {
	fakeAuditLogEntryuery, fakeAuditLogEntryArgs := fakes.BuildFakeSQLQuery()
	mockQueryBuilder.AuditLogEntrySQLQueryBuilder.
		On("BuildCreateAuditLogEntryQuery", mock.IsType(&types.AuditLogEntryCreationInput{})).
		Return(fakeAuditLogEntryuery, fakeAuditLogEntryArgs)

	db.ExpectExec(formatQueryForSQLMock(fakeAuditLogEntryuery)).
		WithArgs(interfaceToDriverValue(fakeAuditLogEntryArgs)...).
		WillReturnResult(newSuccessfulDatabaseResult(123))
}

// end helper funcs

func TestClient_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.ExternalID = ""
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.ExternalID = ""
		exampleAccountCreationInput := &types.AccountCreationInput{
			Name:          exampleUser.Username,
			BelongsToUser: exampleUser.ID,
		}

		exampleInput := &types.TestUserCreationConfig{
			Username:       exampleUser.Username,
			Password:       exampleUser.HashedPassword,
			HashedPassword: exampleUser.HashedPassword,
			IsServiceAdmin: true,
		}

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		// called by c.IsReady()
		db.ExpectPing()

		migrationFuncCalled := false

		// expect BuildMigrationFunc to be called
		mockQueryBuilder.On("BuildMigrationFunc", mock.AnythingOfType("*sql.DB")).
			Return(func() {
				migrationFuncCalled = true
			})

		db.ExpectBegin()

		// expect TestUser to be created
		fakeTestUserCreationQuery, fakeTestUserCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.On("BuildTestUserCreationQuery", exampleInput).
			Return(fakeTestUserCreationQuery, fakeTestUserCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeTestUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeTestUserCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleUser.ID))

		// create audit log entry for created TestUser
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildCreateAuditLogEntryQuery", mock.MatchedBy(testutil.AuditLogEntryCreationInputMatcher(audit.UserCreationEvent))).
			Return(firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs)

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created TestUser
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountSQLQueryBuilder.On("BuildCreateAccountQuery", exampleAccountCreationInput).
			Return(fakeAccountCreationQuery, fakeAccountCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildCreateAuditLogEntryQuery", mock.MatchedBy(testutil.AuditLogEntryCreationInputMatcher(audit.AccountCreationEvent))).
			Return(secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs)

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.On("BuildCreateMembershipForNewUserQuery", exampleUser.ID, exampleAccount.ID).
			Return(fakeMembershipCreationQuery, fakeMembershipCreationArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccount.ID))

		thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AuditLogEntrySQLQueryBuilder.On("BuildCreateAuditLogEntryQuery", mock.MatchedBy(testutil.AuditLogEntryCreationInputMatcher(audit.UserAddedToAccountEvent))).
			Return(thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs)

		db.ExpectExec(formatQueryForSQLMock(thirdFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(thirdFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.NoError(t, c.Migrate(ctx, 1, exampleInput))
		assert.True(t, migrationFuncCalled)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with failure executing queries", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.ExternalID = ""
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleInput := &types.TestUserCreationConfig{
			Username:       exampleUser.Username,
			Password:       exampleUser.HashedPassword,
			HashedPassword: exampleUser.HashedPassword,
			IsServiceAdmin: true,
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		// called by c.IsReady()
		db.ExpectPing()

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		// expect BuildMigrationFunc to be called
		mockQueryBuilder.On("BuildMigrationFunc", mock.AnythingOfType("*sql.DB")).
			Return(func() {})

		// expect TestUser to be created
		fakeTestUserCreationQuery, fakeTestUserCreationArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.On("BuildTestUserCreationQuery", exampleInput).
			Return(fakeTestUserCreationQuery, fakeTestUserCreationArgs)

		c.sqlQueryBuilder = mockQueryBuilder

		// expect transaction begin
		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.Migrate(ctx, 1, exampleInput))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
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

		db.ExpectPing().WillReturnError(errors.New("blah"))

		assert.False(t, c.IsReady(ctx, 1))
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		var migrationFunctionCalled bool
		fakeMigrationFunc := func() {
			migrationFunctionCalled = true
		}

		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		queryBuilder := database.BuildMockSQLQueryBuilder()
		queryBuilder.On("BuildMigrationFunc", mock.AnythingOfType("*sql.DB")).Return(fakeMigrationFunc)

		mockDB.ExpectPing().WillDelayFor(0)

		exampleConfig := &dbconfig.Config{
			Debug:           true,
			RunMigrations:   true,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNonOperationalLogger(), db, exampleConfig, queryBuilder)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.True(t, migrationFunctionCalled)
		mock.AssertExpectationsForObjects(t, &sqlmockExpecterWrapper{Sqlmock: mockDB}, queryBuilder)
	})

	T.Run("with PostgresProviderKey", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		var migrationFunctionCalled bool
		fakeMigrationFunc := func() {
			migrationFunctionCalled = true
		}

		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)

		queryBuilder := database.BuildMockSQLQueryBuilder()
		queryBuilder.On("BuildMigrationFunc", mock.AnythingOfType("*sql.DB")).Return(fakeMigrationFunc)

		mockDB.ExpectPing().WillDelayFor(0)

		exampleConfig := &dbconfig.Config{
			Provider:        dbconfig.PostgresProviderKey,
			Debug:           true,
			RunMigrations:   true,
			MaxPingAttempts: 1,
		}

		actual, err := ProvideDatabaseClient(ctx, logging.NewNonOperationalLogger(), db, exampleConfig, queryBuilder)
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

		actual, err := ProvideDatabaseClient(ctx, logging.NewNonOperationalLogger(), db, exampleConfig, queryBuilder)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, &sqlmockExpecterWrapper{Sqlmock: mockDB}, queryBuilder)
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

	T.Run("handles nil", func(t *testing.T) {
		t.Parallel()

		var c *Client

		assert.NotEmpty(t, c.currentTime())
	})
}

func TestClient_rollbackTransaction(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()
		db.ExpectRollback().WillReturnError(errors.New("blah"))

		tx, err := c.db.BeginTx(ctx, nil)
		require.NoError(t, err)

		c.rollbackTransaction(tx)
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

func TestClient_performCreateQueryIgnoringReturn(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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

func TestClient_performCreateQuery(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
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
}
