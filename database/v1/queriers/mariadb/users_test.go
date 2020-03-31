package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	fake "github.com/brianvoe/gofakeit"
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

func buildFakeUser() *models.User {
	return &models.User{
		ID:             fake.Uint64(),
		Username:       fake.Username(),
		HashedPassword: fake.UUID(),
		// byte arrays don't work with sqlmock
		Salt:            nil,
		TwoFactorSecret: fake.UUID(),
		IsAdmin:         false,
		CreatedOn:       uint64(uint32(fake.Date().Unix())),
	}
}

func buildFakeUserDatabaseCreationInput(user *models.User) models.UserDatabaseCreationInput {
	return models.UserDatabaseCreationInput{
		Username:        user.Username,
		HashedPassword:  user.HashedPassword,
		TwoFactorSecret: user.TwoFactorSecret,
	}
}

func TestMariaDB_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expectedUserID := fake.Uint64()
		expectedArgCount := 1
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = ?"

		actualQuery, args := m.buildGetUserQuery(expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expectedUserID, args[0])
	})
}

func TestMariaDB_GetUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := m.GetUser(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUser(ctx, expected.ID)
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
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL LIMIT 20"

		actualQuery, args := m.buildGetUsersQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_GetUsers(T *testing.T) {
	T.Parallel()

	expectedUsersQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL LIMIT 20"
	expectedCountQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedCount := fake.Uint64()
		expected := &models.UserList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Users: []models.User{
				{
					ID:       fake.Uint64(),
					Username: fake.Username(),
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
		expected := &models.UserList{
			Users: []models.User{
				{
					ID:       fake.Uint64(),
					Username: fake.Username(),
				},
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(&expected.Users[0]))

		actual, err := m.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedCount := fake.Uint64()
		expected := &models.UserList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Users: []models.User{
				{
					ID:       fake.Uint64(),
					Username: fake.Username(),
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

		actual, err := m.GetUsers(ctx, filter)
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
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = ?"

		actualQuery, args := m.buildGetUserByUsernameQuery(expectedUsername)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expectedUsername, args[0])
	})
}

func TestMariaDB_GetUserByUsername(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := m.GetUserByUsername(ctx, expected.Username)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetUserByUsername(ctx, expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUserByUsername(ctx, expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetUserCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		filter := models.DefaultQueryFilter()

		expectedArgCount := 0
		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL LIMIT 20"

		actualQuery, args := m.buildGetUserCountQuery(filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_GetUserCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expected := fake.Uint64()

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := m.GetUserCount(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetUserCount(ctx, filter)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := buildFakeUser()
		expectedInput := buildFakeUserDatabaseCreationInput(expected)
		expectedArgCount := 4
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin,created_on) VALUES (?,?,?,?,UNIX_TIMESTAMP())"

		actualQuery, args := m.buildCreateUserQuery(expectedInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expected.Username, args[0])
		assert.Equal(t, expected.HashedPassword, args[1])
		assert.Equal(t, expected.TwoFactorSecret, args[2])
		assert.Equal(t, expected.IsAdmin, args[3])
	})
}

func TestMariaDB_CreateUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin,created_on) VALUES (?,?,?,?,UNIX_TIMESTAMP())"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()
		expectedInput := buildFakeUserDatabaseCreationInput(expected)
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT users.created_on FROM users WHERE users.id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"users.created_on"}).AddRow(expected.CreatedOn))

		actual, err := m.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()
		expectedInput := buildFakeUserDatabaseCreationInput(expected)

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnError(errors.New("blah"))

		actual, err := m.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := buildFakeUser()
		expectedArgCount := 4
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = UNIX_TIMESTAMP() WHERE id = ?"

		actualQuery, args := m.buildUpdateUserQuery(expected)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expected.Username, args[0])
		assert.Equal(t, expected.HashedPassword, args[1])
		assert.Equal(t, expected.TwoFactorSecret, args[2])
		assert.Equal(t, expected.ID, args[3])
	})
}

func TestMariaDB_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = UNIX_TIMESTAMP() WHERE id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.ID,
		).WillReturnResult(exampleRows)

		err := m.UpdateUser(ctx, expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleUserID := fake.Uint64()
		expectedArgCount := 1
		expectedQuery := "UPDATE users SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE id = ?"

		actualQuery, args := m.buildArchiveUserQuery(exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestMariaDB_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := buildFakeUser()
		expectedQuery := "UPDATE users SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveUser(ctx, expected.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
