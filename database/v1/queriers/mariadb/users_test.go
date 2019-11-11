package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowFromUser(user *models.User) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(usersTableColumns).AddRow(
		user.ID,
		user.Username,
		user.HashedPassword,
		user.PasswordLastChangedOn,
		user.TwoFactorSecret,
		user.IsAdmin,
		user.CreatedOn,
		user.UpdatedOn,
		user.ArchivedOn,
	)

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

func TestMariaDB_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expectedUserID := uint64(123)
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = ?"

		actualQuery, args := m.buildGetUserQuery(expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestMariaDB_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = ?"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := m.GetUser(context.Background(), expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = ?"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUser(context.Background(), expected.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := m.buildGetUsersQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedCountQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"
		expectedCount := uint64(321)
		expected := &models.UserList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Users: []models.User{
				{
					ID:       123,
					Username: "username",
				},
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := m.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"
		expected := &models.UserList{
			Users: []models.User{
				{
					ID:       123,
					Username: "username",
				},
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(&expected.Users[0]))

		actual, err := m.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		expectedCountQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"
		expectedCount := uint64(321)
		expected := &models.UserList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Users: []models.User{
				{
					ID:       123,
					Username: "username",
				},
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		expectedUsername := "username"
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = ?"

		actualQuery, args := m.buildGetUserByUsernameQuery(expectedUsername)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUsername, args[0].(string))
	})
}

func TestMariaDB_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = ?"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := m.GetUserByUsername(context.Background(), expected.Username)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = ?"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUserByUsername(context.Background(), expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = ?"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUserByUsername(context.Background(), expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUserCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := m.buildGetUserCountQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_GetUserCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(123)
		expectedQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := m.GetUserCount(context.Background(), models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUserCount(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := &models.UserInput{
			Username:        "username",
			Password:        "hashed password",
			TwoFactorSecret: "two factor secret",
		}
		expectedArgCount := 4
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin,created_on) VALUES (?,?,?,?,UNIX_TIMESTAMP())"

		actualQuery, args := m.buildCreateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.UserInput{
			Username: expected.Username,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin,created_on) VALUES (?,?,?,?,UNIX_TIMESTAMP())"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT created_on FROM users WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := m.CreateUser(context.Background(), expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.UserInput{
			Username: expected.Username,
		}
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin,created_on) VALUES (?,?,?,?,UNIX_TIMESTAMP())"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnError(errors.New("blah"))

		actual, err := m.CreateUser(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUser := &models.User{
			ID:              321,
			Username:        "username",
			HashedPassword:  "hashed password",
			TwoFactorSecret: "two factor secret",
		}
		expectedArgCount := 4
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = UNIX_TIMESTAMP() WHERE id = ?"

		actualQuery, args := m.buildUpdateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = UNIX_TIMESTAMP() WHERE id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.ID,
		).WillReturnResult(exampleRows)

		err := m.UpdateUser(context.Background(), expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUserID := uint64(321)
		expectedArgCount := 1
		expectedQuery := "UPDATE users SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE id = ?"

		actualQuery, args := m.buildArchiveUserQuery(exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestMariaDB_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE users SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveUser(context.Background(), expected.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
