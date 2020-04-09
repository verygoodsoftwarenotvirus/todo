package mariadb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromUser(users ...*models.User) *sqlmock.Rows {
	includeCount := len(users) > 1
	columns := usersTableColumns

	if includeCount {
		columns = append(columns, "count")
	}
	exampleRows := sqlmock.NewRows(columns)

	for _, user := range users {
		rowValues := []driver.Value{
			user.ID,
			user.Username,
			user.HashedPassword,
			user.PasswordLastChangedOn,
			user.TwoFactorSecret,
			user.IsAdmin,
			user.CreatedOn,
			user.UpdatedOn,
			user.ArchivedOn,
		}

		if includeCount {
			rowValues = append(rowValues, len(users))
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
		user.PasswordLastChangedOn,
		user.TwoFactorSecret,
		user.IsAdmin,
		user.CreatedOn,
		user.UpdatedOn,
	)

	return exampleRows
}

func TestSqlite_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}

		actualQuery, actualArgs := m.buildGetUserQuery(exampleUser.ID)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := m.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on, COUNT(users.id) FROM users WHERE users.archived_on IS NULL AND users.created_on > ? AND users.created_on < ? AND users.updated_on > ? AND users.updated_on < ? GROUP BY users.id LIMIT 20 OFFSET 180"
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

func TestSqlite_GetUsers(T *testing.T) {
	T.Parallel()

	expectedUsersQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on, COUNT(users.id) FROM users WHERE users.archived_on IS NULL GROUP BY users.id LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUser := fakemodels.BuildFakeUser()
		expected := &models.UserList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: 3,
			},
			Users: []models.User{
				*exampleUser,
				*exampleUser,
				*exampleUser,
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowsFromUser(
				exampleUser,
				exampleUser,
				exampleUser,
			),
		)

		actual, err := m.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(exampleUser))

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = ?"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}

		actualQuery, actualArgs := m.buildGetUserByUsernameQuery(exampleUser.Username)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetUserByUsername(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.Username).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := m.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.Username).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.Username).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllUserCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"

		actualQuery := m.buildGetAllUserCountQuery()
		ensureArgCountMatchesQuery(t, actualQuery, nil)
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES (?,?,?,?)"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.IsAdmin,
		}

		actualQuery, actualArgs := m.buildCreateUserQuery(exampleInput)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_CreateUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES (?,?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			false,
		).WillReturnResult(exampleRows)

		mtt := &mockTimeTeller{}
		mtt.On("Now").Return(exampleUser.CreatedOn)
		m.timeTeller = mtt

		actual, err := m.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		m, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			false,
		).WillReturnError(errors.New("blah"))

		actual, err := m.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		}

		actualQuery, actualArgs := m.buildUpdateUserQuery(exampleUser)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()
		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = UNIX_TIMESTAMP() WHERE id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		).WillReturnResult(exampleRows)

		err := m.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE id = ?"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}

		actualQuery, actualArgs := m.buildArchiveUserQuery(exampleUser.ID)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()
		expectedQuery := "UPDATE users SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
