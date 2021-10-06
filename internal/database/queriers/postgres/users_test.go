package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func buildMockRowsFromUsers(includeCounts bool, filteredCount uint64, users ...*types.User) *sqlmock.Rows {
	columns := usersTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, user := range users {
		rowValues := []driver.Value{
			user.ID,
			user.Username,
			user.AvatarSrc,
			user.HashedPassword,
			user.RequiresPasswordChange,
			user.PasswordLastChangedOn,
			user.TwoFactorSecret,
			user.TwoFactorSecretVerifiedOn,
			strings.Join(user.ServiceRoles, serviceRolesSeparator),
			user.ServiceAccountStatus,
			user.ReputationExplanation,
			user.CreatedOn,
			user.LastUpdatedOn,
			user.ArchivedOn,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(users))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanUsers(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On(
			"Next",
		).Return(false)
		mockRows.On(
			"Err",
		).Return(errors.New("blah"))

		_, _, _, err := q.scanUsers(ctx, mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On(
			"Next",
		).Return(false)
		mockRows.On(
			"Err",
		).Return(nil)
		mockRows.On(
			"Close",
		).Return(errors.New("blah"))

		_, _, _, err := q.scanUsers(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_UserHasStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, db := buildTestClient(t)
		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleStatus := string(types.GoodStandingAccountStatus)

		args := []interface{}{exampleUserID, exampleStatus}

		db.ExpectQuery(formatQueryForSQLMock(userHasStatusQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		actual, err := c.UserHasStatus(ctx, exampleUserID, exampleStatus)
		assert.NoError(t, err)
		assert.True(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)
		ctx := context.Background()
		exampleStatus := string(types.GoodStandingAccountStatus)

		actual, err := c.UserHasStatus(ctx, "", exampleStatus)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with empty statuses list", func(t *testing.T) {
		t.Parallel()

		c, _ := buildTestClient(t)
		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()

		actual, err := c.UserHasStatus(ctx, exampleUserID)
		assert.NoError(t, err)
		assert.True(t, actual)
	})

	T.Run("with error performing query", func(t *testing.T) {
		t.Parallel()

		c, db := buildTestClient(t)
		ctx := context.Background()
		exampleUserID := fakes.BuildFakeID()
		exampleStatus := string(types.GoodStandingAccountStatus)

		args := []interface{}{exampleUserID, exampleStatus}

		db.ExpectQuery(formatQueryForSQLMock(userHasStatusQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.UserHasStatus(ctx, exampleUserID, exampleStatus)
		assert.Error(t, err)
		assert.False(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_getUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.getUser(ctx, exampleUser.ID, true)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.getUser(ctx, "", true)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("without verified two factor secret", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserWithUnverifiedTwoFactorQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.getUser(ctx, exampleUser.ID, false)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.getUser(ctx, exampleUser.ID, true)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetUser(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.ID}

		db.ExpectQuery(formatQueryForSQLMock(getUserWithUnverifiedTwoFactorQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetUserWithUnverifiedTwoFactorSecret(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.Username}

		db.ExpectQuery(formatQueryForSQLMock(getUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUser))

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetUserByUsername(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.Username}

		db.ExpectQuery(formatQueryForSQLMock(getUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUser.Username}

		db.ExpectQuery(formatQueryForSQLMock(getUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_SearchForUsersByUsername(T *testing.T) {
	T.Parallel()

	exampleUsername := fakes.BuildFakeUser().Username

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserList := fakes.BuildFakeUserList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			fmt.Sprintf("%s%%", exampleUsername),
		}

		db.ExpectQuery(formatQueryForSQLMock(searchForUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(false, 0, exampleUserList.Users...))

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList.Users, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.SearchForUsersByUsername(ctx, "")
		assert.Error(t, err)
		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			fmt.Sprintf("%s%%", exampleUsername),
		}

		db.ExpectQuery(formatQueryForSQLMock(searchForUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			fmt.Sprintf("%s%%", exampleUsername),
		}

		db.ExpectQuery(formatQueryForSQLMock(searchForUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			fmt.Sprintf("%s%%", exampleUsername),
		}

		db.ExpectQuery(formatQueryForSQLMock(searchForUserByUsernameQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectQuery(formatQueryForSQLMock(getAllUsersCountQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectQuery(formatQueryForSQLMock(getAllUsersCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAllUsersCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t)
	})
}

func TestQuerier_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserList := fakes.BuildFakeUserList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"users",
			nil,
			nil,
			"",
			usersTableColumns,
			"",
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(true, exampleUserList.FilteredCount, exampleUserList.Users...))

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleUserList := fakes.BuildFakeUserList()
		exampleUserList.Limit, exampleUserList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"users",
			nil,
			nil,
			"",
			usersTableColumns,
			"",
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromUsers(true, exampleUserList.FilteredCount, exampleUserList.Users...))

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"users",
			nil,
			nil,
			"",
			usersTableColumns,
			"",
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"users",
			nil,
			nil,
			"",
			usersTableColumns,
			"",
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetUsers(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_createUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create account for created user
		accountCreationInput := types.AccountCreationInputForNewUser(exampleUser)
		accountCreationInput.ID = exampleAccount.ID
		accountCreationArgs := []interface{}{
			accountCreationInput.ID,
			accountCreationInput.Name,
			types.UnpaidAccountBillingStatus,
			accountCreationInput.ContactEmail,
			accountCreationInput.ContactPhone,
			accountCreationInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		// create account user membership for created user
		createAccountMembershipForNewUserArgs := []interface{}{
			&idMatcher{},
			exampleUser.ID,
			exampleAccount.ID,
			true,
			authorization.AccountAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(createAccountMembershipForNewUserQuery)).
			WithArgs(interfaceToDriverValue(createAccountMembershipForNewUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		assert.NoError(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		assert.Error(t, c.createUser(ctx, &types.User{}, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		query, args := fakes.BuildFakeSQLQuery()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, query, args))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating user in database", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating account", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create account for created user
		accountCreationInput := types.AccountCreationInputForNewUser(exampleUser)
		accountCreationInput.ID = exampleAccount.ID
		accountCreationArgs := []interface{}{
			accountCreationInput.ID,
			accountCreationInput.Name,
			types.UnpaidAccountBillingStatus,
			accountCreationInput.ContactEmail,
			accountCreationInput.ContactPhone,
			accountCreationInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error creating account user membership", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create account for created user
		accountCreationInput := types.AccountCreationInputForNewUser(exampleUser)
		accountCreationInput.ID = exampleAccount.ID
		accountCreationArgs := []interface{}{
			accountCreationInput.ID,
			accountCreationInput.Name,
			types.UnpaidAccountBillingStatus,
			accountCreationInput.ContactEmail,
			accountCreationInput.ContactPhone,
			accountCreationInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		// create account user membership for created user

		createAccountMembershipForNewUserArgs := []interface{}{
			&idMatcher{},
			exampleUser.ID,
			exampleAccount.ID,
			true,
			authorization.AccountAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(createAccountMembershipForNewUserQuery)).
			WithArgs(interfaceToDriverValue(createAccountMembershipForNewUserArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		fakeUserCreationQuery, fakeUserCreationArgs := fakes.BuildFakeSQLQuery()

		db.ExpectExec(formatQueryForSQLMock(fakeUserCreationQuery)).
			WithArgs(interfaceToDriverValue(fakeUserCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create account for created user
		accountCreationInput := types.AccountCreationInputForNewUser(exampleUser)
		accountCreationInput.ID = exampleAccount.ID
		accountCreationArgs := []interface{}{
			accountCreationInput.ID,
			accountCreationInput.Name,
			types.UnpaidAccountBillingStatus,
			accountCreationInput.ContactEmail,
			accountCreationInput.ContactPhone,
			accountCreationInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		// create account user membership for created user

		createAccountMembershipForNewUserArgs := []interface{}{
			&idMatcher{},
			exampleUser.ID,
			exampleAccount.ID,
			true,
			authorization.AccountAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(createAccountMembershipForNewUserQuery)).
			WithArgs(interfaceToDriverValue(createAccountMembershipForNewUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.createUser(ctx, exampleUser, exampleAccount, fakeUserCreationQuery, fakeUserCreationArgs))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.TwoFactorSecretVerifiedOn = nil
		exampleUser.CreatedOn = exampleCreationTime
		exampleUser.ServiceAccountStatus = ""
		exampleUserCreationInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		exampleAccount := fakes.BuildFakeAccountForUser(exampleUser)
		exampleAccount.CreatedOn = exampleCreationTime

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		db.ExpectBegin()

		userCreationArgs := []interface{}{
			exampleUserCreationInput.ID,
			exampleUserCreationInput.Username,
			exampleUserCreationInput.HashedPassword,
			exampleUserCreationInput.TwoFactorSecret,
			types.UnverifiedAccountStatus,
			authorization.ServiceUserRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(userCreationQuery)).
			WithArgs(interfaceToDriverValue(userCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		// create account for created user
		accountCreationInput := types.AccountCreationInputForNewUser(exampleUser)
		accountCreationInput.ID = exampleAccount.ID
		accountCreationArgs := []interface{}{
			&idMatcher{},
			accountCreationInput.Name,
			types.UnpaidAccountBillingStatus,
			accountCreationInput.ContactEmail,
			accountCreationInput.ContactPhone,
			&idMatcher{},
		}

		db.ExpectExec(formatQueryForSQLMock(accountCreationQuery)).
			WithArgs(interfaceToDriverValue(accountCreationArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		// create account user membership for created user
		createAccountMembershipForNewUserArgs := []interface{}{
			&idMatcher{},
			&idMatcher{},
			&idMatcher{},
			true,
			authorization.AccountAdminRole.String(),
		}

		db.ExpectExec(formatQueryForSQLMock(createAccountMembershipForNewUserQuery)).
			WithArgs(interfaceToDriverValue(createAccountMembershipForNewUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAccount.ID))

		db.ExpectCommit()

		actual, err := c.CreateUser(ctx, exampleUserCreationInput)
		assert.NoError(t, err)
		actual.ID = exampleUser.ID
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateUser(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error creating user", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fakes.BuildFakeTime()

		exampleUser := fakes.BuildFakeUser()
		exampleUserCreationInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		ctx := context.Background()
		c, db := buildTestClient(t)

		c.timeFunc = func() uint64 {
			return exampleCreationTime
		}

		begin := db.ExpectBegin()
		begin.WillReturnError(errors.New("blah"))

		actual, err := c.CreateUser(ctx, exampleUserCreationInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.AvatarSrc,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		assert.NoError(t, c.UpdateUser(ctx, exampleUser))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUser(ctx, nil))
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleUser.Username,
			exampleUser.HashedPassword,
			exampleUser.AvatarSrc,
			exampleUser.TwoFactorSecret,
			exampleUser.TwoFactorSecretVerifiedOn,
			exampleUser.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateUserQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUser(ctx, exampleUser))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleHashedPassword := "$argon2i$v=19$m=64,t=10,p=4$RjFtMmRmU2lGYU9CMk1mMw$cuGR9AhTczPR6xDOSAMW+SvEYFyLEIS+7nlRdC9f6ys"

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleHashedPassword,
			false,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateUserPasswordQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUserID))

		assert.NoError(t, c.UpdateUserPassword(ctx, exampleUserID, exampleHashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleHashedPassword := "$argon2i$v=19$m=64,t=10,p=4$RjFtMmRmU2lGYU9CMk1mMw$cuGR9AhTczPR6xDOSAMW+SvEYFyLEIS+7nlRdC9f6ys"

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserPassword(ctx, "", exampleHashedPassword))
	})

	T.Run("with invalid new hash", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, ""))
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleHashedPassword := "$argon2i$v=19$m=64,t=10,p=4$RjFtMmRmU2lGYU9CMk1mMw$cuGR9AhTczPR6xDOSAMW+SvEYFyLEIS+7nlRdC9f6ys"

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleHashedPassword,
			false,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateUserPasswordQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUserID, exampleHashedPassword))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_UpdateUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			nil,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateUserTwoFactorSecretQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		assert.NoError(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, "", exampleUser.TwoFactorSecret))
	})

	T.Run("with invalid new secret", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, ""))
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			nil,
			exampleUser.TwoFactorSecret,
			exampleUser.ID,
		}

		db.ExpectExec(formatQueryForSQLMock(updateUserTwoFactorSecretQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, exampleUser.TwoFactorSecret))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_VerifyUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			types.GoodStandingAccountStatus,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(markUserTwoFactorSecretAsVerified)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, ""))
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			types.GoodStandingAccountStatus,
			exampleUserID,
		}

		db.ExpectExec(formatQueryForSQLMock(markUserTwoFactorSecretAsVerified)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		archiveUserArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveUserQuery)).
			WithArgs(interfaceToDriverValue(archiveUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUserID))

		archiveMembershipsArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(archiveMembershipsArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUserID))

		db.ExpectCommit()

		assert.NoError(t, c.ArchiveUser(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveUser(ctx, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveUser(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing user archive query", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		archiveUserArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveUserQuery)).
			WithArgs(interfaceToDriverValue(archiveUserArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveUser(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing memberships archive query", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		archiveUserArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveUserQuery)).
			WithArgs(interfaceToDriverValue(archiveUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUserID))

		archiveMembershipsArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(archiveMembershipsArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		assert.Error(t, c.ArchiveUser(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		archiveUserArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveUserQuery)).
			WithArgs(interfaceToDriverValue(archiveUserArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUserID))

		archiveMembershipsArgs := []interface{}{exampleUserID}

		db.ExpectExec(formatQueryForSQLMock(archiveMembershipsQuery)).
			WithArgs(interfaceToDriverValue(archiveMembershipsArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUserID))

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveUser(ctx, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})
}
