package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	postgres "github.com/lib/pq"
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
			user.RequiresPasswordChange,
			user.PasswordLastChangedOn,
			user.TwoFactorSecret,
			user.IsAdmin,
			user.TwoFactorSecretVerifiedOn,
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
		user.RequiresPasswordChange,
		user.PasswordLastChangedOn,
		user.TwoFactorSecret,
		user.IsAdmin,
		user.TwoFactorSecretVerifiedOn,
		user.CreatedOn,
		user.UpdatedOn,
	)

	return exampleRows
}

func TestPostgres_ScanUsers(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := p.scanUsers(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := p.scanUsers(mockRows)
		assert.NoError(t, err)
	})
}

func TestPostgres_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildGetUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = $1 AND users.two_factor_secret_verified_on IS NOT NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := p.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = $1 AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.id = $1 AND users.two_factor_secret_verified_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := p.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.ID,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
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

		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on, COUNT(users.id) FROM users WHERE users.archived_on IS NULL AND users.created_on > $1 AND users.created_on < $2 AND users.updated_on > $3 AND users.updated_on < $4 GROUP BY users.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := p.buildGetUsersQuery(filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetUsers(T *testing.T) {
	T.Parallel()

	expectedUsersQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on, COUNT(users.id) FROM users WHERE users.archived_on IS NULL GROUP BY users.id LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		exampleUserList := fakemodels.BuildFakeUserList()
		exampleUserList.Users[0].Salt = nil
		exampleUserList.Users[1].Salt = nil
		exampleUserList.Users[2].Salt = nil

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).WillReturnRows(
			buildMockRowsFromUser(
				&exampleUserList.Users[0],
				&exampleUserList.Users[1],
				&exampleUserList.Users[2],
			),
		)

		actual, err := p.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()

		filter := models.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedUsersQuery)).
			WillReturnRows(buildErroneousMockRowFromUser(fakemodels.BuildFakeUser()))

		actual, err := p.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := p.buildGetUserByUsernameQuery(exampleUser.Username)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetUserByUsername(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.is_admin, users.two_factor_secret_verified_on, users.created_on, users.updated_on, users.archived_on FROM users WHERE users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.Salt = nil

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.Username,
			).
			WillReturnRows(buildMockRowsFromUser(exampleUser))

		actual, err := p.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.Username,
			).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(
				exampleUser.Username,
			).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllUserCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := p.buildGetAllUsersCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllUserCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleCount := uint64(123)

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := p.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAllUsersCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.IsAdmin,
		}
		actualQuery, actualArgs := p.buildCreateUserQuery(exampleInput)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreateUser(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret,is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_on"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.Salt = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleUser.ID, exampleUser.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			false,
		).WillReturnRows(exampleRows)

		actual, err := p.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with postgres row exists error", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			false,
		).WillReturnError(&postgres.Error{Code: postgresRowExistsErrorCode})

		actual, err := p.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, err, dbclient.ErrUserExists)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			false,
		).WillReturnError(errors.New("blah"))

		actual, err := p.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, two_factor_secret = $3, two_factor_secret_verified_on = $4, updated_on = extract(epoch FROM NOW()) WHERE id = $5 RETURNING updated_on"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildUpdateUserQuery(exampleUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, two_factor_secret = $3, two_factor_secret_verified_on = $4, updated_on = extract(epoch FROM NOW()) WHERE id = $5 RETURNING updated_on"
		exampleRows := sqlmock.NewRows([]string{"updated_on"}).AddRow(uint64(time.Now().Unix()))

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		).WillReturnRows(exampleRows)

		err := p.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), updated_on = extract(epoch FROM NOW()) WHERE id = $3"
		expectedArgs := []interface{}{
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		expectedQuery := "UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), updated_on = extract(epoch FROM NOW()) WHERE id = $3"

		p, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		).WillReturnResult(driver.ResultNoRows)

		err := p.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()) WHERE id = $1"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_VerifyUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()) WHERE id = $1"

		p, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.VerifyUserTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		exampleUser := fakemodels.BuildFakeUser()

		expectedQuery := "UPDATE users SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildArchiveUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		expectedQuery := "UPDATE users SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUser.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
