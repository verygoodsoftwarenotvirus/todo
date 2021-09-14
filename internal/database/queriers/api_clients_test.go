package queriers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAPIClients(includeCounts bool, filteredCount uint64, clients ...*types.APIClient) *sqlmock.Rows {
	columns := querybuilding.APIClientsTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, c := range clients {
		rowValues := []driver.Value{
			c.ID,
			c.Name,
			c.ClientID,
			c.ClientSecret,
			c.CreatedOn,
			c.LastUpdatedOn,
			c.ArchivedOn,
			c.BelongsToUser,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(clients))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanAPIClients(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAPIClients(ctx, mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanAPIClients(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetAPIClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientByClientIDQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ClientID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(false, 0, exampleAPIClient))

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with empty client ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAPIClientByClientID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientByClientIDQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ClientID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientByClientIDQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ClientID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAPIClientByDatabaseID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientByDatabaseIDQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(false, 0, exampleAPIClient))

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClient.ID, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAPIClientByDatabaseID(ctx, "", exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClient.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientByDatabaseIDQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClient.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientByDatabaseIDQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			exampleUser.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClient.ID, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetTotalAPIClientCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAllAPIClientsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetTotalAPIClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAllAPIClientsCountQuery",
			testutils.ContextMatcher,
		).Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetTotalAPIClientCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAPIClients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientsQuery",
			testutils.ContextMatcher,
			exampleUser.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(true, exampleAPIClientList.FilteredCount, exampleAPIClientList.Clients...))

		actual, err := c.GetAPIClients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClientList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAPIClients(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		exampleAPIClientList.Limit, exampleAPIClientList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientsQuery",
			testutils.ContextMatcher,
			exampleUser.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(true, exampleAPIClientList.FilteredCount, exampleAPIClientList.Clients...))

		actual, err := c.GetAPIClients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClientList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientsQuery",
			testutils.ContextMatcher,
			exampleUser.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetAPIClients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientsQuery",
			testutils.ContextMatcher,
			exampleUser.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAPIClients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUser := fakes.BuildFakeUser()
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAPIClientsQuery",
			testutils.ContextMatcher,
			exampleUser.ID, filter,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAPIClients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

var apiClientCreationInputMatcher interface{} = mock.MatchedBy(func(input *types.APIClientCreationInput) bool {
	return true
})

func TestQuerier_CreateAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
		exampleAPIClient.ID = exampleInput.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildCreateAPIClientQuery",
			testutils.ContextMatcher,
			apiClientCreationInputMatcher,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAPIClient(ctx, exampleInput, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid creator ID", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAPIClient(ctx, exampleInput, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAPIClient(ctx, nil, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		actual, err := c.CreateAPIClient(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildCreateAPIClientQuery",
			testutils.ContextMatcher,
			apiClientCreationInputMatcher,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAPIClient(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildCreateAPIClientQuery",
			testutils.ContextMatcher,
			apiClientCreationInputMatcher,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAPIClient(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildCreateAPIClientQuery",
			testutils.ContextMatcher,
			apiClientCreationInputMatcher,
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAPIClient(ctx, exampleInput, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_ArchiveAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildArchiveAPIClientQuery",
			testutils.ContextMatcher,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.NoError(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAPIClient(ctx, "", exampleAccount.ID, exampleUser.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, "", exampleUser.ID))
	})

	T.Run("with invalid actor ID", func(t *testing.T) {
		t.Parallel()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAccount.ID, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildArchiveAPIClientQuery",
			testutils.ContextMatcher,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error writing audit log entry to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildArchiveAPIClientQuery",
			testutils.ContextMatcher,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, errors.New("blah"))

		db.ExpectRollback()

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error committing transaction", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildArchiveAPIClientQuery",
			testutils.ContextMatcher,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
		).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db, nil)

		db.ExpectCommit().WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAccount.ID, exampleUser.ID))

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAuditLogEntriesForAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAuditLogEntriesList := fakes.BuildFakeAuditLogEntryList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAPIClientQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(
				false,
				exampleAuditLogEntriesList.Entries...,
			))

		actual, err := c.GetAuditLogEntriesForAPIClient(ctx, exampleAPIClient.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntriesList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAuditLogEntriesForAPIClient(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAPIClientQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForAPIClient(ctx, exampleAPIClient.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On(
			"BuildGetAuditLogEntriesForAPIClientQuery",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
		).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForAPIClient(ctx, exampleAPIClient.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}
