package querier

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromOAuth2Clients(includeCounts bool, filteredCount uint64, clients ...*types.OAuth2Client) *sqlmock.Rows {
	columns := queriers.OAuth2ClientsTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, c := range clients {
		rowValues := []driver.Value{
			c.ID,
			c.Name,
			c.ClientID,
			strings.Join(c.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
			c.RedirectURI,
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

func TestClient_ScanOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanOAuth2Clients(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestClient(t)

		mockRows := &database.MockResultIterator{}
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanOAuth2Clients(mockRows, false)
		assert.Error(t, err)
	})
}

func TestClient_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientQuery", exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, 0, exampleOAuth2Client))

		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientQuery", exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientQuery", exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientByClientIDQuery", exampleOAuth2Client.ClientID).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, 0, exampleOAuth2Client))

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientByClientIDQuery", exampleOAuth2Client.ClientID).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetTotalOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetAllOAuth2ClientsCountQuery").Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetTotalOAuth2ClientCount(ctx)
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
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetAllOAuth2ClientsCountQuery").Return(fakeQuery)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetTotalOAuth2ClientCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.OAuth2Client)
		expectedCount := uint64(20)
		doneChan := make(chan bool, 1)
		exampleBatchSize := uint16(1000)
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAllOAuth2ClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetBatchOfOAuth2ClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, 0, exampleOAuth2ClientList.Clients...))

		err := c.GetAllOAuth2Clients(ctx, results, exampleBatchSize)
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

		results := make(chan []*types.OAuth2Client)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAllOAuth2ClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetBatchOfOAuth2ClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		err := c.GetAllOAuth2Clients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.OAuth2Client)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAllOAuth2ClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		c.sqlQueryBuilder = mockQueryBuilder

		err := c.GetAllOAuth2Clients(ctx, results, exampleBatchSize)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.OAuth2Client)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAllOAuth2ClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetBatchOfOAuth2ClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnError(errors.New("blah"))

		err := c.GetAllOAuth2Clients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.OAuth2Client)
		expectedCount := uint64(20)
		exampleBatchSize := uint16(1000)

		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, _ := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAllOAuth2ClientsCountQuery").
			Return(fakeQuery, []interface{}{})

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expectedCount))

		secondFakeQuery, secondFakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetBatchOfOAuth2ClientsQuery", uint64(1), uint64(exampleBatchSize+1)).
			Return(secondFakeQuery, secondFakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(secondFakeQuery)).
			WithArgs(interfaceToDriverValue(secondFakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		err := c.GetAllOAuth2Clients(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(true, exampleOAuth2ClientList.FilteredCount, exampleOAuth2ClientList.Clients...))

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		exampleOAuth2ClientList.Limit, exampleOAuth2ClientList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(true, exampleOAuth2ClientList.FilteredCount, exampleOAuth2ClientList.Clients...))

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
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
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildGetOAuth2ClientsQuery", exampleUser.ID, filter).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildCreateOAuth2ClientQuery", exampleInput).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleOAuth2Client.ID))

		c.timeFunc = func() uint64 {
			return exampleOAuth2Client.CreatedOn
		}

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildCreateOAuth2ClientQuery", exampleInput).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleOAuth2Client.CreatedOn
		}

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		var expected error
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildUpdateOAuth2ClientQuery", exampleOAuth2Client).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleOAuth2Client.ID))

		actual := c.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		var expected error
		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildArchiveOAuth2ClientQuery", exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleOAuth2Client.ID))

		actual := c.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_LogOAuth2ClientCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleClient := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntry := audit.BuildOAuth2ClientCreationEventEntry(exampleClient)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogOAuth2ClientCreationEvent(ctx, exampleClient)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_LogOAuth2ClientArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleClient := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntry := audit.BuildOAuth2ClientArchiveEventEntry(exampleUser.ID, exampleClient.ID)

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		prepareForAuditLogEntryCreation(t, exampleAuditLogEntry, mockQueryBuilder, db)
		c.sqlQueryBuilder = mockQueryBuilder

		c.LogOAuth2ClientArchiveEvent(ctx, exampleUser.ID, exampleClient.ID)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}

func TestClient_GetAuditLogEntriesForOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntriesList := fakes.BuildFakeAuditLogEntryList()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForOAuth2ClientQuery", exampleOAuth2Client.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(
				false,
				exampleAuditLogEntriesList.Entries...,
			))

		actual, err := c.GetAuditLogEntriesForOAuth2Client(ctx, exampleOAuth2Client.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntriesList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForOAuth2ClientQuery", exampleOAuth2Client.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForOAuth2Client(ctx, exampleOAuth2Client.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ctx := context.Background()
		c, db := buildTestClient(t)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForOAuth2ClientQuery", exampleOAuth2Client.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForOAuth2Client(ctx, exampleOAuth2Client.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}
