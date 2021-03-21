package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

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
			c.ExternalID,
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

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanAPIClients(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errs", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanAPIClients(mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetAPIClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientByClientIDQuery", exampleAPIClient.ClientID).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(false, 0, exampleAPIClient))

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientByClientIDQuery", exampleAPIClient.ClientID).Return(fakeQuery, fakeArgs)
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
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientByClientIDQuery", exampleAPIClient.ClientID).Return(fakeQuery, fakeArgs)
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

func TestQuerier_GetTotalAPIClientCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAllAPIClientsCountQuery").Return(fakeQuery)
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
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAllAPIClientsCountQuery").Return(fakeQuery)
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

func TestQuerier_GetAllAPIClients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.APIClient)
		expectedCount := uint64(20)
		doneChan := make(chan bool, 1)
		exampleBatchSize := uint16(1000)
		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAllAPIClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetBatchOfAPIClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(false, 0, exampleAPIClientList.Clients...))

		err := c.GetAllAPIClients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		var stillQuerying = true
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

		results := make(chan []*types.APIClient)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAllAPIClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetBatchOfAPIClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		err := c.GetAllAPIClients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.APIClient)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAllAPIClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllAPIClients(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.APIClient)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAllAPIClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetBatchOfAPIClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		err := c.GetAllAPIClients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.APIClient)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAllAPIClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetBatchOfAPIClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		err := c.GetAllAPIClients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_GetAPIClients(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAPIClients(true, exampleAPIClientList.FilteredCount, exampleAPIClientList.Clients...))

		actual, err := c.GetAPIClients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClientList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		exampleAPIClientList.Limit, exampleAPIClientList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
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

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
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

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
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

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildGetAPIClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
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

func TestQuerier_CreateAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ExternalID = ""
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildCreateAPIClientQuery", exampleInput).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db)

		db.ExpectCommit()

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAPIClient(ctx, exampleInput, 0)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildCreateAPIClientQuery", exampleInput).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		db.ExpectRollback()

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}
		c.sqlQueryBuilder = mockQueryBuilder

		actual, err := c.CreateAPIClient(ctx, exampleInput, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestQuerier_ArchiveAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		var expected error
		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		db.ExpectBegin()

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.On("BuildArchiveAPIClientQuery", exampleAPIClient.ID, exampleAPIClient.BelongsToUser).Return(fakeQuery, fakeArgs)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleAPIClient.ID))

		expectAuditLogEntryInTransaction(mockQueryBuilder, db)

		db.ExpectCommit()

		c.sqlQueryBuilder = mockQueryBuilder

		actual := c.ArchiveAPIClient(ctx, exampleAPIClient.ID, exampleAPIClient.BelongsToUser, 0)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

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
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAPIClientQuery", exampleAPIClient.ID).
			Return(fakeQuery, fakeArgs)
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

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAPIClientQuery", exampleAPIClient.ID).
			Return(fakeQuery, fakeArgs)
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
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.APIClientSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAPIClientQuery", exampleAPIClient.ID).
			Return(fakeQuery, fakeArgs)
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
