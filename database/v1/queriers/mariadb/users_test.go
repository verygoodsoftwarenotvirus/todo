package mariadb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

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

func TestMariaDB_ScanUsers(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := m.scanUsers(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := m.scanUsers(mockRows)
		assert.NoError(t, err)
	})
}

func TestMariaDB_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = ? AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := m.buildGetUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := m.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = ? AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := m.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := m.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.created_on > ? AND users.created_on < ? AND users.last_updated_on > ? AND users.last_updated_on < ? ORDER BY users.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := m.buildGetUsersQuery(filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		exampleUserList := fakemodels.BuildFakeUserList()
		exampleUserList.Users[0].Salt = nil
		exampleUserList.Users[1].Salt = nil
		exampleUserList.Users[2].Salt = nil

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromUser(
					&exampleUserList.Users[0],
					&exampleUserList.Users[1],
					&exampleUserList.Users[2],
				),
			)

		actual, err := m.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromUser(fakemodels.BuildFakeUser()))

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = ? AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := m.buildGetUserByUsernameQuery(exampleUser.Username)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := m.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllUsersCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := m.buildGetAllUsersCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		m, mockDB := buildTestService(t)
		expectedQuery := m.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := m.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)
		expectedQuery := m.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllUsersCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

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
		actualQuery, actualArgs := m.buildCreateUserQuery(exampleInput)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.Salt = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := m.buildCreateUserQuery(expectedInput)

		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		mtt := &mockTimeTeller{}
		mtt.On("Now").Return(exampleUser.CreatedOn)
		m.timeTeller = mtt

		actual, err := m.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mtt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := m.buildCreateUserQuery(expectedInput)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := m.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, salt = ?, two_factor_secret = ?, two_factor_secret_verified_on = ?, last_updated_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}
		actualQuery, actualArgs := m.buildUpdateUserQuery(exampleUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildUpdateUserQuery(exampleUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		err := m.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = ?, requires_password_change = ?, password_last_changed_on = UNIX_TIMESTAMP(), last_updated_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		}
		actualQuery, actualArgs := m.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(driver.ResultNoRows)

		err := m.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := m.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_VerifyUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.VerifyUserTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := m.buildArchiveUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := m.buildArchiveUserQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
