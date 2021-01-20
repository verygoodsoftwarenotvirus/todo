package superclient

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

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

func buildErroneousMockRowFromOAuth2Client(c *types.OAuth2Client) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.OAuth2ClientsTableColumns).AddRow(
		c.ArchivedOn,
		c.Name,
		c.ClientID,
		strings.Join(c.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
		c.RedirectURI,
		c.ClientSecret,
		c.CreatedOn,
		c.LastUpdatedOn,
		c.BelongsToUser,
		c.ID,
	)

	return exampleRows
}

func TestSqlite_ScanOAuth2Clients(T *testing.T) {
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
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

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
}

func TestClient_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

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
}

func TestClient_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

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
}

func TestClient_GetOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, db := buildTestClient(t)
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		filter := types.DefaultQueryFilter()

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
		ctx := context.Background()

		c, db := buildTestClient(t)
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		exampleOAuth2ClientList.Limit, exampleOAuth2ClientList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

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
}

func TestClient_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, db := buildTestClient(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.OAuth2ClientSQLQueryBuilder.On("BuildCreateOAuth2ClientQuery", mock.MatchedBy(matchOAuth2Client(t, exampleOAuth2Client))).Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newSuccessfulDatabaseResult(exampleOAuth2Client.ID))

		mtt := &queriers.MockTimeTeller{}
		mtt.On("Now").Return(exampleOAuth2Client.CreatedOn)
		c.timeTeller = mtt

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder, mtt)
	})
}

func TestClient_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		var expected error
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
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		var expected error
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
