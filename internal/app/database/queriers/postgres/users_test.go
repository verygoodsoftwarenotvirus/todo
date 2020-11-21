package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	postgres "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromUsers(users ...*types.User) *sqlmock.Rows {
	columns := queriers.UsersTableColumns
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
			user.AdminPermissions,
			user.AccountStatus,
			user.AccountStatusExplanation,
			user.CreatedOn,
			user.LastUpdatedOn,
			user.ArchivedOn,
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromUser(user *types.User) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.UsersTableColumns).AddRow(
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
		user.AdminPermissions,
		user.AccountStatus,
		user.AccountStatusExplanation,
		user.CreatedOn,
		user.LastUpdatedOn,
	)

	return exampleRows
}

func TestPostgres_ScanUsers(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := p.scanUsers(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := p.scanUsers(mockRows)
		assert.NoError(t, err)
	})
}

func TestPostgres_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.admin_permissions, users.account_status, users.status_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUsers(exampleUser))

		actual, err := p.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserWithUnverifiedTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.admin_permissions, users.account_status, users.status_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NULL"
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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUsers(exampleUser))

		actual, err := p.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.admin_permissions, users.account_status, users.status_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.created_on > $1 AND users.created_on < $2 AND users.last_updated_on > $3 AND users.last_updated_on < $4 ORDER BY users.id LIMIT 20 OFFSET 180"
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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		exampleUserList := fakes.BuildFakeUserList()
		exampleUserList.Users[0].Salt = nil
		exampleUserList.Users[1].Salt = nil
		exampleUserList.Users[2].Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromUsers(
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
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromUser(fakes.BuildFakeUser()))

		actual, err := p.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.admin_permissions, users.account_status, users.status_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUsers(exampleUser))

		actual, err := p.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildSearchForUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_admin, users.admin_permissions, users.account_status, users.status_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.username ILIKE $1 AND users.archived_on IS NULL AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			fmt.Sprintf("%s%%", exampleUser.Username),
		}
		actualQuery, actualArgs := p.buildSearchForUserByUsernameQuery(exampleUser.Username)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_SearchForUsersByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUsername := fakes.BuildFakeUser().Username
		exampleUsers := fakes.BuildFakeUserList().Users
		exampleUsers[0].Salt = nil
		exampleUsers[1].Salt = nil
		exampleUsers[2].Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildSearchForUserByUsernameQuery(exampleUsername)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromUsers(
					&exampleUsers[0],
					&exampleUsers[1],
					&exampleUsers[2],
				),
			)

		actual, err := p.SearchForUsersByUsername(ctx, exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, exampleUsers, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUsername := fakes.BuildFakeUser().Username
		exampleUsers := fakes.BuildFakeUserList().Users
		exampleUsers[0].Salt = nil
		exampleUsers[1].Salt = nil
		exampleUsers[2].Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildSearchForUserByUsernameQuery(exampleUsername)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUsername := fakes.BuildFakeUser().Username
		exampleUsers := fakes.BuildFakeUserList().Users
		exampleUsers[0].Salt = nil
		exampleUsers[1].Salt = nil
		exampleUsers[2].Salt = nil

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildSearchForUserByUsernameQuery(exampleUsername)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllUsersCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := p.buildGetAllUsersCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := p.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
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
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery := "INSERT INTO users (username,hashed_password,salt,two_factor_secret,account_status,is_admin,admin_permissions) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			types.UnverifiedStandingAccountStatus,
			false,
			0,
		}
		actualQuery, actualArgs := p.buildCreateUserQuery(exampleInput)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.Salt = nil
		expectedInput := fakes.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := p.buildCreateUserQuery(expectedInput)

		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleUser.ID, exampleUser.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		actual, err := p.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with postgres row exists error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakes.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := p.buildCreateUserQuery(expectedInput)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(&postgres.Error{Code: postgresRowExistsErrorCode})

		actual, err := p.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, dbclient.ErrUserExists))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakes.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := p.buildCreateUserQuery(expectedInput)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, salt = $3, two_factor_secret = $4, two_factor_secret_verified_on = $5, last_updated_on = extract(epoch FROM NOW()) WHERE id = $6 RETURNING last_updated_on"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
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
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleRows := sqlmock.NewRows([]string{"last_updated_on"}).AddRow(uint64(time.Now().Unix()))

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildUpdateUserQuery(exampleUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		err := p.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), last_updated_on = extract(epoch FROM NOW()) WHERE id = $3 RETURNING last_updated_on"
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
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(driver.ResultNoRows)

		err := p.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()), account_status = $1 WHERE id = $2"
		expectedArgs := []interface{}{
			types.GoodStandingAccountStatus,
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
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.VerifyUserTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildBanUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET account_status = $1 WHERE id = $2"
		expectedArgs := []interface{}{
			types.BannedStandingAccountStatus,
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildBanUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BanUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildBanUserQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.BanUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"
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
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildArchiveUserQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAuditLogEntriesForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE (audit_log.context->'user_id' = $1 OR audit_log.context->'performed_by' = $2) ORDER BY audit_log.id"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := p.buildGetAuditLogEntriesForUserQuery(exampleUser.ID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_LogUserCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserCreationEventEntry(exampleInput)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUserCreationEvent(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserVerifyTwoFactorSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserVerifyTwoFactorSecretEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleUser.ID, exampleUser.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUserVerifyTwoFactorSecretEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserUpdateTwoFactorSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserUpdateTwoFactorSecretEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUserUpdateTwoFactorSecretEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserUpdatePasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserUpdatePasswordEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUserUpdatePasswordEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserArchiveEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUserArchiveEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogSuccessfulLoginEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildSuccessfulLoginEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogSuccessfulLoginEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUnsuccessfulLoginBadPasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUnsuccessfulLoginBadPasswordEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUnsuccessfulLoginBad2FATokenEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogUnsuccessfulLoginBad2FATokenEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogLogoutEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildLogoutEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogLogoutEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
