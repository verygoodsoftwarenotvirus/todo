package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAccountUserMemberships(includeCounts bool, filteredCount uint64, memberships ...*types.AccountUserMembership) *sqlmock.Rows {
	columns := querybuilding.AccountsUserMembershipTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, x := range memberships {
		rowValues := []driver.Value{
			x.ID,
			x.ExternalID,
			x.BelongsToUser,
			x.BelongsToAccount,
			x.UserPermissions,
			x.CreatedOn,
			x.ArchivedOn,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(memberships))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestClient_ScanAccountUserMemberships(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountUserMemberships(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanAccountUserMemberships(mockRows, false)
		assert.Error(t, err)
	})
}

func TestClient_GetAccountUserMembership(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipQuery", exampleAccountUserMembership.ID, exampleUser.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(false, 0, exampleAccountUserMembership))

		actual, err := c.GetAccountUserMembership(ctx, exampleAccountUserMembership.ID, exampleAccountUserMembership.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountUserMembership, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipQuery", exampleAccountUserMembership.ID, exampleUser.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccountUserMembership(ctx, exampleAccountUserMembership.ID, exampleAccountUserMembership.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllAccountUserMembershipsCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAllAccountUserMembershipsCountQuery").
			Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(uint64(123)))

		actual, err := c.GetAllAccountUserMembershipsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllAccountUserMemberships(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		results := make(chan []*types.AccountUserMembership)
		doneChan := make(chan bool, 1)
		expectedCount := uint64(20)
		exampleAccountUserMembershipList := fakes.BuildFakeAccountUserMembershipList()
		exampleBatchSize := uint16(1000)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAllAccountUserMembershipsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetBatchOfAccountUserMembershipsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(false, 0, exampleAccountUserMembershipList.AccountUserMemberships...))

		err := c.GetAllAccountUserMemberships(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		stillQuerying := true
		for stillQuerying {
			select {
			case batch := <-results:
				assert.NotEmpty(t, batch)
				doneChan <- true
			case <-time.After(time.Second):
				t.FailNow()
			case <-doneChan:
				stillQuerying = false
			}
		}

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.AccountUserMembership)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAllAccountUserMembershipsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetBatchOfAccountUserMembershipsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		err := c.GetAllAccountUserMemberships(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.AccountUserMembership)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAllAccountUserMembershipsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllAccountUserMemberships(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.AccountUserMembership)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAllAccountUserMembershipsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetBatchOfAccountUserMembershipsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		err := c.GetAllAccountUserMemberships(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.AccountUserMembership)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAllAccountUserMembershipsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetBatchOfAccountUserMembershipsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		err := c.GetAllAccountUserMemberships(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAccountUserMemberships(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountUserMembershipList := fakes.BuildFakeAccountUserMembershipList()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", exampleUser.ID, false, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(
				true,
				exampleAccountUserMembershipList.FilteredCount,
				exampleAccountUserMembershipList.AccountUserMemberships...,
			))

		actual, err := c.GetAccountUserMemberships(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountUserMembershipList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccountUserMembershipList := fakes.BuildFakeAccountUserMembershipList()
		exampleAccountUserMembershipList.Page = 0
		exampleAccountUserMembershipList.Limit = 0

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", exampleUser.ID, false, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(
				true,
				exampleAccountUserMembershipList.FilteredCount,
				exampleAccountUserMembershipList.AccountUserMemberships...,
			))

		actual, err := c.GetAccountUserMemberships(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountUserMembershipList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", exampleUser.ID, false, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccountUserMemberships(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", exampleUser.ID, false, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccountUserMemberships(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAccountUserMembershipsForAdmin(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleAccountUserMembershipList := fakes.BuildFakeAccountUserMembershipList()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", uint64(0), true, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(
				true,
				exampleAccountUserMembershipList.FilteredCount,
				exampleAccountUserMembershipList.AccountUserMemberships...,
			))

		actual, err := c.GetAccountUserMembershipsForAdmin(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountUserMembershipList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		filter := (*types.QueryFilter)(nil)
		exampleAccountUserMembershipList := fakes.BuildFakeAccountUserMembershipList()
		exampleAccountUserMembershipList.Page = 0
		exampleAccountUserMembershipList.Limit = 0

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", uint64(0), true, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAccountUserMemberships(
				true,
				exampleAccountUserMembershipList.FilteredCount,
				exampleAccountUserMembershipList.AccountUserMemberships...,
			))

		actual, err := c.GetAccountUserMembershipsForAdmin(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountUserMembershipList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", uint64(0), true, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAccountUserMembershipsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAccountUserMembershipsQuery", uint64(0), true, filter).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAccountUserMembershipsForAdmin(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_CreateAccountUserMembership(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.ExternalID = ""
		exampleInput := fakes.BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership(exampleAccountUserMembership)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildCreateAccountUserMembershipQuery", exampleInput).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountUserMembership.ID))

		c.timeFunc = func() uint64 {
			return exampleAccountUserMembership.CreatedOn
		}

		actual, err := c.CreateAccountUserMembership(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountUserMembership, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleInput := fakes.BuildFakeAccountUserMembershipCreationInputFromAccountUserMembership(exampleAccountUserMembership)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildCreateAccountUserMembershipQuery", exampleInput).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAccountUserMembership.CreatedOn
		}

		actual, err := c.CreateAccountUserMembership(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_ArchiveAccountUserMembership(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAccountUserMembership.BelongsToUser = exampleUser.ID

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildArchiveAccountUserMembershipQuery", exampleAccountUserMembership.ID, exampleAccountUserMembership.BelongsToUser).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAccountUserMembership.ID))

		err := c.ArchiveAccountUserMembership(ctx, exampleAccountUserMembership.ID, exampleAccountUserMembership.BelongsToUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAuditLogEntriesForAccountUserMembership(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAuditLogEntriesList := fakes.BuildFakeAuditLogEntryList()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAccountUserMembershipQuery", exampleAccountUserMembership.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(
				false,
				exampleAuditLogEntriesList.Entries...,
			))

		actual, err := c.GetAuditLogEntriesForAccountUserMembership(ctx, exampleAccountUserMembership.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntriesList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAccountUserMembershipQuery", exampleAccountUserMembership.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForAccountUserMembership(ctx, exampleAccountUserMembership.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAccountUserMembershipQuery", exampleAccountUserMembership.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForAccountUserMembership(ctx, exampleAccountUserMembership.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}
