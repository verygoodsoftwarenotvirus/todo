package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	postgres "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromUsers(includeCount bool, users ...*types.User) *sqlmock.Rows {
	columns := queriers.UsersTableColumns

	if includeCount {
		columns = append(columns, "count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, user := range users {
		rowValues := []driver.Value{
			user.ID,
			user.Username,
			user.AvatarSrc,
			user.HashedPassword,
			user.Salt,
			user.RequiresPasswordChange,
			user.PasswordLastChangedOn,
			user.TwoFactorSecret,
			user.TwoFactorSecretVerifiedOn,
			user.IsSiteAdmin,
			user.AdminPermissions,
			user.AccountStatus,
			user.AccountStatusExplanation,
			user.CreatedOn,
			user.LastUpdatedOn,
			user.ArchivedOn,
		}

		if includeCount {
			rowValues = append(rowValues, len(users))
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
		user.AvatarSrc,
		user.HashedPassword,
		user.Salt,
		user.RequiresPasswordChange,
		user.PasswordLastChangedOn,
		user.TwoFactorSecret,
		user.TwoFactorSecretVerifiedOn,
		user.IsSiteAdmin,
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
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := q.scanUsers(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := q.scanUsers(mockRows, false)
		assert.NoError(t, err)
	})
}

func TestPostgres_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_site_admin, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.buildGetUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, exampleUser))

		actual, err := q.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetUser(ctx, exampleUser.ID)
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
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_site_admin, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.id = $1 AND users.two_factor_secret_verified_on IS NULL"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, exampleUser))

		actual, err := q.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserWithUnverifiedTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
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
		q, _ := buildTestService(t)

		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_site_admin, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on, (SELECT COUNT(*) FROM users WHERE users.archived_on IS NULL AND items.created_on > $1 AND items.created_on < $2 AND items.last_updated_on > $3 AND items.last_updated_on < $4) FROM users WHERE users.archived_on IS NULL AND users.created_on > $5 AND users.created_on < $6 AND users.last_updated_on > $7 AND users.last_updated_on < $8 ORDER BY users.created_on LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.buildGetUsersQuery(filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromUsers(
					true,
					&exampleUserList.Users[0],
					&exampleUserList.Users[1],
					&exampleUserList.Users[2],
				),
			)

		actual, err := q.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUsersQuery(filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromUser(fakes.BuildFakeUser()))

		actual, err := q.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_site_admin, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			exampleUser.Username,
		}
		actualQuery, actualArgs := q.buildGetUserByUsernameQuery(exampleUser.Username)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromUsers(false, exampleUser))

		actual, err := q.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildGetUserByUsernameQuery(exampleUser.Username)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildSearchForUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUsername := fakes.BuildFakeUser().Username

		expectedQuery := "SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.salt, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.is_site_admin, users.site_admin_permissions, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.username ILIKE $1 AND users.archived_on IS NULL AND users.two_factor_secret_verified_on IS NOT NULL"
		expectedArgs := []interface{}{
			fmt.Sprintf("%s%%", exampleUsername),
		}
		actualQuery, actualArgs := q.buildSearchForUserByUsernameQuery(exampleUsername)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildSearchForUserByUsernameQuery(exampleUsername)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromUsers(
					false,
					&exampleUsers[0],
					&exampleUsers[1],
					&exampleUsers[2],
				),
			)

		actual, err := q.SearchForUsersByUsername(ctx, exampleUsername)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildSearchForUserByUsernameQuery(exampleUsername)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.SearchForUsersByUsername(ctx, exampleUsername)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildSearchForUserByUsernameQuery(exampleUsername)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.SearchForUsersByUsername(ctx, exampleUsername)
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
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(users.id) FROM users WHERE users.archived_on IS NULL"
		actualQuery := q.buildGetAllUsersCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		q, mockDB := buildTestService(t)
		expectedQuery := q.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := q.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		expectedQuery := q.buildGetAllUsersCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := q.GetAllUsersCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		expectedQuery := "INSERT INTO users (username,hashed_password,salt,two_factor_secret,reputation,is_site_admin,site_admin_permissions) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleUser.TwoFactorSecret,
			types.UnverifiedAccountStatus,
			false,
			0,
		}
		actualQuery, actualArgs := q.buildCreateUserQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.Salt = nil
		expectedInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := q.buildCreateUserQuery(expectedInput)

		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleUser.ID, exampleUser.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		actual, err := q.CreateUser(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with postgres row exists error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := q.buildCreateUserQuery(expectedInput)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(&postgres.Error{Code: postgresRowExistsErrorCode})

		actual, err := q.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, dbclient.ErrUserExists))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		expectedInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		expectedQuery, expectedArgs := q.buildCreateUserQuery(expectedInput)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.CreateUser(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

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
		actualQuery, actualArgs := q.buildUpdateUserQuery(exampleUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildUpdateUserQuery(exampleUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		err := q.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateUserPasswordQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET hashed_password = $1, requires_password_change = $2, password_last_changed_on = extract(epoch FROM NOW()), last_updated_on = extract(epoch FROM NOW()) WHERE id = $3 RETURNING last_updated_on"
		expectedArgs := []interface{}{
			exampleUser.HashedPassword,
			false,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildUpdateUserPasswordQuery(exampleUser.ID, exampleUser.HashedPassword)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(driver.ResultNoRows)

		err := q.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildVerifyUserTwoFactorSecretQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET two_factor_secret_verified_on = extract(epoch FROM NOW()), reputation = $1 WHERE id = $2"
		expectedArgs := []interface{}{
			types.GoodStandingAccountStatus,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildVerifyUserTwoFactorSecretQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := q.VerifyUserTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "UPDATE users SET archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"
		expectedArgs := []interface{}{
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.buildArchiveUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.buildArchiveUserQuery(exampleUser.ID)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := q.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAuditLogEntriesForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE (audit_log.context->'user_id' = $1 OR audit_log.context->'performed_by' = $2) ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			exampleUser.ID,
			exampleUser.ID,
		}
		actualQuery, actualArgs := q.buildGetAuditLogEntriesForUserQuery(exampleUser.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_LogUserCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserCreationEventEntry(exampleInput)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUserCreationEvent(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserVerifyTwoFactorSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserVerifyTwoFactorSecretEventEntry(exampleUser.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleUser.ID, exampleUser.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUserVerifyTwoFactorSecretEvent(ctx, exampleUser.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserUpdateTwoFactorSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserUpdateTwoFactorSecretEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUserUpdateTwoFactorSecretEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserUpdatePasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserUpdatePasswordEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUserUpdatePasswordEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUserArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUserArchiveEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUserArchiveEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogSuccessfulLoginEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildSuccessfulLoginEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogSuccessfulLoginEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUnsuccessfulLoginBadPasswordEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUnsuccessfulLoginBadPasswordEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUnsuccessfulLoginBadPasswordEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogUnsuccessfulLoginBad2FATokenEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogUnsuccessfulLoginBad2FATokenEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogLogoutEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeUser()
		exampleAuditLogEntryInput := audit.BuildLogoutEventEntry(exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := q.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogLogoutEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
