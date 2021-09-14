package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromUsers(includeCounts bool, filteredCount uint64, users ...*types.User) *sqlmock.Rows {
	columns := querybuilding.UsersTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, user := range users {
		rowValues := []driver.Value{
			user.ID,
			user.Username,
			user.AvatarSrc,
			user.HashedPassword,
			user.RequiresPasswordChange,
			user.PasswordLastChangedOn,
			user.TwoFactorSecret,
			user.TwoFactorSecretVerifiedOn,
			strings.Join(user.ServiceRoles, serviceRolesSeparator),
			user.ServiceAccountStatus,
			user.ReputationExplanation,
			user.CreatedOn,
			user.LastUpdatedOn,
			user.ArchivedOn,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(users))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanUsers(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On(
			"Next",
		).Return(false)
		mockRows.On(
			"Err",
		).Return(errors.New("blah"))

		_, _, _, err := q.scanUsers(ctx, mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On(
			"Next",
		).Return(false)
		mockRows.On(
			"Err",
		).Return(nil)
		mockRows.On(
			"Close",
		).Return(errors.New("blah"))

		_, _, _, err := q.scanUsers(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_UserHasStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, db := buildTestClient(t)
		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleStatus := string(types.GoodStandingAccountStatus)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := c.UserHasStatus(ctx, exampleUser.ID, exampleStatus)
		assert.NoError(t, err)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)
		ctx := context.Background()
		exampleStatus := string(types.GoodStandingAccountStatus)

		actual, err := c.UserHasStatus(ctx, "", exampleStatus)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with empty statuses list", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)
		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()

		actual, err := c.UserHasStatus(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.True(t, actual)
	})

	T.Run("with error performing query", func(t *testing.T) {
		t.Parallel()

		c, db := buildTestClient(t)
		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleStatus := string(types.GoodStandingAccountStatus)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.UserHasStatus(ctx, exampleUser.ID, exampleStatus)
		assert.Error(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_getUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.getUser(ctx, exampleUser.ID, true)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.getUser(ctx, "", true)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("without verified two factor secret", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.getUser(ctx, exampleUser.ID, false)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.getUser(ctx, exampleUser.ID, true)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetUser(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetUserWithUnverifiedTwoFactorSecret(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetUserByUsername(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_SearchForUsersByUsername(T *testing.T) {
	T.Parallel()

	exampleUsername := fakes.BuildFakeUser().Username

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserList := fakes.BuildFakeUserList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUserList.Users...))

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList.Users, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.SearchForUsersByUsername(ctx, "")
		assert.Error(t, err)
		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAllUsersCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t)
	})
}

func TestQuerier_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserList := fakes.BuildFakeUserList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(true, exampleUserList.FilteredCount, exampleUserList.Users...))

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleUserList := fakes.BuildFakeUserList()
		exampleUserList.Limit, exampleUserList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromUsers(true, exampleUserList.FilteredCount, exampleUserList.Users...))

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_createUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(thirdFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(thirdFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit()

		assert.NoError(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		query, args := fakes.BuildFakeSQLQuery()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, query, args))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating user in database", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing user creation audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnError(errors.New("blahy"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating account", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing account creation audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating account user membership", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing account membership creation audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(thirdFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(thirdFakeAuditLogEntryEventArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(thirdFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(thirdFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})
}

// accountCreationInputMatcher is a matcher for use with testify/mock's MatchBy function.
func accountCreationInputMatcher(t *testing.T, expected *types.AccountCreationInput) func(*types.AccountCreationInput) bool {
	t.Helper()

	return func(actual *types.AccountCreationInput) bool {
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.ContactEmail, actual.ContactEmail)
		assert.Equal(t, expected.ContactPhone, actual.ContactPhone)
		assert.NotEmpty(t, actual.BelongsToUser)

		return true
	}
}

// userDataStoreCreationInputMatcher is a matcher for use with testify/mock's MatchBy function.
func userDataStoreCreationInputMatcher(t *testing.T, expected *types.UserDataStoreCreationInput) func(*types.UserDataStoreCreationInput) bool {
	t.Helper()

	return func(actual *types.UserDataStoreCreationInput) bool {
		assert.Equal(t, expected.Username, actual.Username)
		assert.Equal(t, expected.HashedPassword, actual.HashedPassword)
		assert.Equal(t, expected.TwoFactorSecret, actual.TwoFactorSecret)

		return true
	}
}

func TestQuerier_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime
		exampleUser.ServiceAccountStatus = ""
		exampleUserCreationInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime
		//exampleAccountCreationInput := types.AccountCreationInputForNewUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create audit log entry for created user
		firstFakeAuditLogEntryEventQuery, firstFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(firstFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(firstFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account for created user
		fakeAccountCreationQuery, fakeAccountCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeAccountCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeAccountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		secondFakeAuditLogEntryEventQuery, secondFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(secondFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(secondFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// create account user membership for created user
		fakeMembershipCreationQuery, fakeMembershipCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeMembershipCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeMembershipCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		thirdFakeAuditLogEntryEventQuery, thirdFakeAuditLogEntryEventArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(thirdFakeAuditLogEntryEventQuery)).
			WithArgs(interfaceToDriverValue(thirdFakeAuditLogEntryEventArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit()

		actual, err := c.CreateUser(ctx, exampleUserCreationInput)
		assert.NoError(t, err)
		actual.ID = exampleUser.ID
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateUser(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error creating user", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUserCreationInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		begin := db.ExpectBegin()
		begin.WillReturnError(errors.New("blah"))

		actual, err := c.CreateUser(ctx, exampleUserCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit()

		assert.NoError(t, c.UpdateUser(ctx, exampleUser, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUser(ctx, nil, nil))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUser(ctx, exampleUser, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateUser(ctx, exampleUser, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectRollback()

		assert.Error(t, c.UpdateUser(ctx, exampleUser, nil))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUser(ctx, exampleUser, nil))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.HashedPassword = "$2b$10$3euPcmQFCiblsZeEu5s7p.9OVHgeHWFDk9nhMqZ0m/3pd/lhwZgES"

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit()

		assert.NoError(t, c.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with issue beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.HashedPassword = "$2b$10$3euPcmQFCiblsZeEu5s7p.9OVHgeHWFDk9nhMqZ0m/3pd/lhwZgES"

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleHashedPassword := "$2b$10$3euPcmQFCiblsZeEu5s7p.9OVHgeHWFDk9nhMqZ0m/3pd/lhwZgES"

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserPassword(ctx, "", exampleHashedPassword))
	})

	T.Run("with invalid new hash", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, ""))
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.HashedPassword = "$2b$10$3euPcmQFCiblsZeEu5s7p.9OVHgeHWFDk9nhMqZ0m/3pd/lhwZgES"

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.HashedPassword = "$2b$10$3euPcmQFCiblsZeEu5s7p.9OVHgeHWFDk9nhMqZ0m/3pd/lhwZgES"

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectRollback()

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.HashedPassword = "$2b$10$3euPcmQFCiblsZeEu5s7p.9OVHgeHWFDk9nhMqZ0m/3pd/lhwZgES"

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit()

		assert.NoError(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, "", exampleUser.TwoFactorSecret))
	})

	T.Run("with invalid new secret", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectRollback()

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_VerifyUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit()

		assert.NoError(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectRollback()

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeArchiveQuery, fakeArchiveArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		fakeArchiveMembershipsQuery, fakeArchiveMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveMembershipsArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit()

		assert.NoError(t, c.ArchiveUser(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveUser(ctx, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveUser(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing user archive query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveUser(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing memberships archive query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeArchiveQuery, fakeArchiveArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		fakeArchiveMembershipsQuery, fakeArchiveMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveMembershipsArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveUser(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing user archive audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeArchiveQuery, fakeArchiveArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		fakeArchiveMembershipsQuery, fakeArchiveMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveMembershipsArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveUser(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeArchiveQuery, fakeArchiveArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		fakeArchiveMembershipsQuery, fakeArchiveMembershipsArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeArchiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(fakeArchiveMembershipsArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveUser(ctx, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAuditLogEntriesForUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(false, exampleAuditLogEntryList.Entries...))

		actual, err := c.GetAuditLogEntriesForUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntriesForUser(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}
