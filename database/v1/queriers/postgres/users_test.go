package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	postgres "github.com/lib/pq"
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

func TestPostgres_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedUserID := uint64(123)
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = $1"

		actualQuery, args := p.buildGetUserQuery(expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestPostgres_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = $1"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := p.GetUser(context.Background(), expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE id = $1"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUser(context.Background(), expected.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := p.buildGetUsersQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_GetUsers(T *testing.T) {
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

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := p.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedUsersQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUsers(context.Background(), models.DefaultQueryFilter())
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

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(&expected.Users[0]))

		actual, err := p.GetUsers(context.Background(), models.DefaultQueryFilter())
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

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
			buildMockRowFromUser(&expected.Users[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedUsername := "username"
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = $1"

		actualQuery, args := p.buildGetUserByUsernameQuery(expectedUsername)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUsername, args[0].(string))
	})
}

func TestPostgres_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = $1"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnRows(buildMockRowFromUser(expected))

		actual, err := p.GetUserByUsername(context.Background(), expected.Username)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = $1"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUserByUsername(context.Background(), expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, is_admin, created_on, updated_on, archived_on FROM users WHERE username = $1"
		expected := &models.User{
			ID:       123,
			Username: "username",
		}

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.Username).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUserByUsername(context.Background(), expected.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := p.buildGetUserCountQuery(models.DefaultQueryFilter())
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_GetUserCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(123)
		expectedQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := p.GetUserCount(context.Background(), models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM users WHERE archived_on IS NULL LIMIT 20"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUserCount(context.Background(), models.DefaultQueryFilter())
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUser := &models.UserInput{
			Username:        "username",
			Password:        "hashed password",
			TwoFactorSecret: "two factor secret",
		}
		expectedArgCount := 4
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_on"

		actualQuery, args := p.buildCreateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_CreateUser(T *testing.T) {
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
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(expected.ID, uint64(time.Now().Unix()))
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnRows(exampleRows)

		actual, err := p.CreateUser(context.Background(), expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with postgres row exists error", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.UserInput{
			Username: expected.Username,
		}
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnError(&postgres.Error{
			Code: postgres.ErrorCode("23505"),
		})

		actual, err := p.CreateUser(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, err, dbclient.ErrUserExists)

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
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.IsAdmin,
		).WillReturnError(errors.New("blah"))

		actual, err := p.CreateUser(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUser := &models.User{
			ID:              321,
			Username:        "username",
			HashedPassword:  "hashed password",
			TwoFactorSecret: "two factor secret",
		}
		expectedArgCount := 4
		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, two_factor_secret = $3, updated_on = extract(epoch FROM NOW()) WHERE id = $4 RETURNING updated_on"

		actualQuery, args := p.buildUpdateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		exampleRows := sqlmock.NewRows([]string{"updated_on"}).AddRow(uint64(time.Now().Unix()))
		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, two_factor_secret = $3, updated_on = extract(epoch FROM NOW()) WHERE id = $4 RETURNING updated_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Username,
			expected.HashedPassword,
			expected.TwoFactorSecret,
			expected.ID,
		).WillReturnRows(exampleRows)

		err := p.UpdateUser(context.Background(), expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUserID := uint64(321)
		expectedArgCount := 1
		expectedQuery := "UPDATE users SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"

		actualQuery, args := p.buildArchiveUserQuery(exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestPostgres_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.User{
			ID:        123,
			Username:  "username",
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE users SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.ArchiveUser(context.Background(), expected.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
