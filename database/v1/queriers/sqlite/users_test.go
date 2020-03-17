package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

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

func TestSqlite_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedUserID := fake.Uint64()
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = ?"

		actualQuery, args := s.buildGetUserQuery(expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestSqlite_GetUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:       fake.Uint64(),
			Username: fake.Username(),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := s.GetUser(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:       fake.Uint64(),
			Username: fake.Username(),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUser(ctx, expected.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := s.buildGetUsersQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_GetUsers(T *testing.T) {
	T.Parallel()

	expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedCountQuery := "SELECT COUNT(users.id) FROM users WHERE archived_on IS NULL LIMIT 20"
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

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := s.GetUsers(ctx, models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUsers(ctx, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUsers(ctx, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.UserList{
			Users: []models.User{
				{
					ID:       fake.Uint64(),
					Username: fake.Username(),
				},
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(&expected.Users[0]))

		actual, err := s.GetUsers(ctx, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		ctx := context.Background()
		expectedCountQuery := "SELECT COUNT(users.id) FROM users WHERE archived_on IS NULL LIMIT 20"
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

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUsers(ctx, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)

		expectedUsername := "username"
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = ?"

		actualQuery, args := s.buildGetUserByUsernameQuery(expectedUsername)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUsername, args[0].(string))
	})
}

func TestSqlite_GetUserByUsername(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:       fake.Uint64(),
			Username: fake.Username(),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := s.GetUserByUsername(ctx, expected.Username)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:       fake.Uint64(),
			Username: fake.Username(),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUserByUsername(ctx, expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:       fake.Uint64(),
			Username: fake.Username(),
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUserByUsername(ctx, expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUserCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := s.buildGetUserCountQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_GetUserCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(users.id) FROM users WHERE archived_on IS NULL LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := s.GetUserCount(ctx, models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUserCount(ctx, models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleUser := &models.UserInput{
			Username:        fake.Username(),
			Password:        "hashed password",
			TwoFactorSecret: "two factor secret",
		}
		expectedArgCount := 4
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES (?,?,?,?)"

		actualQuery, args := s.buildCreateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_CreateUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES (?,?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:        fake.Uint64(),
			Username:  fake.Username(),
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.UserInput{
			Username: expected.Username,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT created_on FROM users WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := s.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:        fake.Uint64(),
			Username:  fake.Username(),
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.UserInput{
			Username: expected.Username,
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnError(errors.New("blah"))

		actual, err := s.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleUser := &models.User{
			ID:              fake.Uint64(),
			Username:        fake.Username(),
			HashedPassword:  "hashed password",
			TwoFactorSecret: "two factor secret",
		}
		expectedArgCount := 4
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = (strftime('%s','now')) WHERE id = ?"

		actualQuery, args := s.buildUpdateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:        fake.Uint64(),
			Username:  fake.Username(),
			CreatedOn: uint64(time.Now().Unix()),
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = (strftime('%s','now')) WHERE id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.ID,
		).WillReturnResult(exampleRows)

		err := s.UpdateUser(ctx, expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleUserID := fake.Uint64()
		expectedArgCount := 1
		expectedQuery := "UPDATE users SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE id = ?"

		actualQuery, args := s.buildArchiveUserQuery(exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestSqlite_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.User{
			ID:        fake.Uint64(),
			Username:  fake.Username(),
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE users SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveUser(ctx, expected.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
