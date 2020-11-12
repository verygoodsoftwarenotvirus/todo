package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/converters"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromUser(users ...*models.User) *sqlmock.Rows {
	columns := usersTableColumns
	exampleRows := sqlmock.NewRows(columns)

	for _, user := range users {
		rowValues := []driver.Value{
			user.ID,
			user.Username,
			user.HashedPassword,
			user.Salt,
			user.RequiresPasswordChange,
			user.PasswordLastChangedOn,
			user.TwoFactorSecret,
			user.TwoFactorSecretVerifiedOn,
			user.IsAdmin,
			user.CreatedOn,
			user.LastUpdatedOn,
			user.ArchivedOn,
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromUser(user *models.User) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(usersTableColumns).AddRow(
		user.ArchivedOn,
		user.ID,
		user.Username,
		user.HashedPassword,
		user.Salt,
		user.RequiresPasswordChange,
		user.PasswordLastChangedOn,
		user.TwoFactorSecret,
		user.TwoFactorSecretVerifiedOn,
		user.IsAdmin,
		user.CreatedOn,
		user.LastUpdatedOn,
	)

	return exampleRows
}

func TestSqlite_ScanUsers(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := s.scanUsers(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := s.scanUsers(mockRows)
		assert.NoError(t, err)
	})
}

func TestSqlite_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = ? AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := s.buildGetUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := s.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = ? AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := s.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := s.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.created_on > ? AND users.created_on < ? AND users.last_updated_on > ? AND users.last_updated_on < ? ORDER BY users.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := s.buildGetUsersQuery(filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		exampleUserList := fakemodels.BuildFakeUserList()
		exampleUserList.Users[0].Salt = nil
		exampleUserList.Users[1].Salt = nil
		exampleUserList.Users[2].Salt = nil

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromUser(
					&exampleUserList.Users[0],
					&exampleUserList.Users[1],
					&exampleUserList.Users[2],
				),
			)

		actual, err := s.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromUser(fakemodels.BuildFakeUser()))

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = ? AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := s.buildGetUserByUsernameQuery(exampleUser.Username)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := s.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllUsersCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := s.buildGetAllUsersCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := s.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllUsersCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery := "INSERT INTO users (username,hashed_password,salt,two_factor_secret,is_admin) VALUES (?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			exampleUser.IsAdmin,
		}
		actualQuery, actualArgs := s.buildCreateUserQuery(exampleInput)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.Salt = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := s.buildCreateUserQuery(expectedInput)

		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		stt := &mockTimeTeller{}
		stt.On("Now").Return(exampleUser.CreatedOn)
		s.timeTeller = stt

		actual, err := s.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, stt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := s.buildCreateUserQuery(expectedInput)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, salt = ?, two_factor_secret = ?, two_factor_secret_verified_on = ?, last_updated_on = (strftime('%s','now')) WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}
		actualQuery, actualArgs := s.buildUpdateUserQuery(exampleUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildUpdateUserQuery(exampleUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		err := s.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = ?, requires_password_change = ?, password_last_changed_on = (strftime('%s','now')), last_updated_on = (strftime('%s','now')) WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		}
		actualQuery, actualArgs := s.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(driver.ResultNoRows)

		err := s.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = (strftime('%s','now')) WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := s.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_VerifyUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.VerifyUserTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = (strftime('%s','now')) WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := s.buildArchiveUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildArchiveUserQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUserCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserCreationEventEntry(exampleUser)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUserCreationEvent(ctx, exampleUser)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUserVerifyTwoFactorSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserVerifyTwoFactorSecretEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUserVerifyTwoFactorSecretEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUserUpdateTwoFactorSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserUpdateTwoFactorSecretEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUserUpdateTwoFactorSecretEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUserUpdatePasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleInput := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserUpdatePasswordEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUserUpdatePasswordEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogUserArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserArchiveEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogUserArchiveEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
