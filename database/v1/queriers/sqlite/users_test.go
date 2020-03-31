package sqlite

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

func buildUserDatabaseCreationInput(user *models.User) models.UserDatabaseCreationInput {
	return models.UserDatabaseCreationInput{
		Username:        user.Username,
		HashedPassword:  user.HashedPassword,
		TwoFactorSecret: user.TwoFactorSecret,
	}
}

func TestSqlite_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedUserID := fake.Uint64()
		expectedArgCount := 1
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = ?"

		actualQuery, args := s.buildGetUserQuery(expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expectedUserID, args[0])
	})
}

func TestSqlite_GetUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(buildMockRowFromUser(exampleUser))

		actual, err := s.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUser(ctx, exampleUser.ID)
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
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL LIMIT 20"

		actualQuery, args := s.buildGetUsersQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_GetUsers(T *testing.T) {
	T.Parallel()

	expectedUsersQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL LIMIT 20"
	expectedCountQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUser := buildFakeUser()
		expected := &models.UserList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: fake.Uint64(),
			},
			Users: []models.User{
				*exampleUser,
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(exampleUser),
			buildMockRowFromUser(exampleUser),
			buildMockRowFromUser(exampleUser),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected.TotalCount))

		actual, err := s.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(exampleUser))

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(exampleUser),
			buildMockRowFromUser(exampleUser),
			buildMockRowFromUser(exampleUser),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)

		expectedUsername := fake.Username()
		expectedArgCount := 1
		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = ?"

		actualQuery, args := s.buildGetUserByUsernameQuery(expectedUsername)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expectedUsername, args[0])
	})
}

func TestSqlite_GetUserByUsername(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.Username).
			WillReturnRows(buildMockRowFromUser(exampleUser))

		actual, err := s.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.Username).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.Username).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUserByUsername(ctx, exampleUser.Username)
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
		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL LIMIT 20"

		actualQuery, args := s.buildGetUserCountQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_GetUserCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expected := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := s.GetUserCount(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetUserCount(ctx, filter)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleUser := buildFakeUser()
		expectedInput := buildUserDatabaseCreationInput(exampleUser)
		expectedArgCount := 4
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES (?,?,?,?)"

		actualQuery, args := s.buildCreateUserQuery(expectedInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUser.Username, args[0])
		assert.Equal(t, exampleUser.HashedPassword, args[1])
		assert.Equal(t, exampleUser.TwoFactorSecret, args[2])
		assert.Equal(t, false, args[3])
	})
}

func TestSqlite_CreateUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES (?,?,?,?)"
	expectedTimeQuery := "SELECT users.created_on FROM users WHERE users.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()
		expectedInput := buildUserDatabaseCreationInput(exampleUser)
		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.IsAdmin,
		).WillReturnResult(exampleRows)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"users.created_on"}).AddRow(exampleUser.CreatedOn))

		actual, err := s.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()
		expectedInput := buildUserDatabaseCreationInput(exampleUser)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.IsAdmin,
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
		exampleUser := buildFakeUser()
		expectedArgCount := 4
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = (strftime('%s','now')) WHERE id = ?"

		actualQuery, args := s.buildUpdateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUser.Username, args[0])
		assert.Equal(t, exampleUser.HashedPassword, args[1])
		assert.Equal(t, exampleUser.TwoFactorSecret, args[2])
		assert.Equal(t, exampleUser.ID, args[3])
	})
}

func TestSqlite_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()
		exampleRows := sqlmock.NewResult(int64(exampleUser.ID), 1)
		expectedQuery := "UPDATE users SET username = ?, hashed_password = ?, two_factor_secret = ?, updated_on = (strftime('%s','now')) WHERE id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		).WillReturnResult(exampleRows)

		err := s.UpdateUser(ctx, exampleUser)
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

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestSqlite_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := buildFakeUser()
		expectedQuery := "UPDATE users SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
