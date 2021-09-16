package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromAPIClients(includeCounts bool, filteredCount uint64, clients ...*types.APIClient) *sqlmock.Rows {
	columns := apiClientsTableColumns

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

		args := []interface{}{exampleAPIClient.ClientID}

		db.ExpectQuery(formatQueryForSQLMock(getAPIClientByClientIDQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAPIClients(false, 0, exampleAPIClient))

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		args := []interface{}{exampleAPIClient.ClientID}

		db.ExpectQuery(formatQueryForSQLMock(getAPIClientByClientIDQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleAPIClient.ClientID}

		db.ExpectQuery(formatQueryForSQLMock(getAPIClientByClientIDQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAPIClientByClientID(ctx, exampleAPIClient.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAPIClientByDatabaseID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClient := fakes.BuildFakeAPIClient()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, exampleAPIClient.ID}

		db.ExpectQuery(formatQueryForSQLMock(getAPIClientByDatabaseIDQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAPIClients(false, 0, exampleAPIClient))

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClient.ID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetAPIClientByDatabaseID(ctx, "", exampleUserID)
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

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClientID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, exampleAPIClientID}

		db.ExpectQuery(formatQueryForSQLMock(getAPIClientByDatabaseIDQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClientID, exampleUserID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClientID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, exampleAPIClientID}

		db.ExpectQuery(formatQueryForSQLMock(getAPIClientByDatabaseIDQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAPIClientByDatabaseID(ctx, exampleAPIClientID, exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetTotalAPIClientCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCount := uint64(123)

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectQuery(formatQueryForSQLMock(getTotalAPIClientCountQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(exampleCount))

		actual, err := c.GetTotalAPIClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectQuery(formatQueryForSQLMock(getTotalAPIClientCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := c.GetTotalAPIClientCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAPIClients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"api_clients",
			nil,
			nil,
			userOwnershipColumn,
			apiClientsTableColumns,
			exampleUserID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAPIClients(true, exampleAPIClientList.FilteredCount, exampleAPIClientList.Clients...))

		actual, err := c.GetAPIClients(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClientList, actual)

		mock.AssertExpectationsForObjects(t, db)
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

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		exampleAPIClientList.Limit, exampleAPIClientList.Page = 0, 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)
		query, args := c.buildListQuery(
			ctx,
			"api_clients",
			nil,
			nil,
			userOwnershipColumn,
			apiClientsTableColumns,
			exampleUserID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromAPIClients(true, exampleAPIClientList.FilteredCount, exampleAPIClientList.Clients...))

		actual, err := c.GetAPIClients(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClientList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("respects sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUserID := fakes.BuildFakeID()
		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"api_clients",
			nil,
			nil,
			userOwnershipColumn,
			apiClientsTableColumns,
			exampleUserID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := c.GetAPIClients(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUserID := fakes.BuildFakeID()
		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"api_clients",
			nil,
			nil,
			userOwnershipColumn,
			apiClientsTableColumns,
			exampleUserID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAPIClients(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()
		exampleUserID := fakes.BuildFakeID()
		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"api_clients",
			nil,
			nil,
			userOwnershipColumn,
			apiClientsTableColumns,
			exampleUserID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAPIClients(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_CreateAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)
		exampleAPIClient.ID = exampleInput.ID

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			exampleInput.ClientID,
			exampleInput.ClientSecret,
			exampleInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(createAPIClientQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClient.ID))

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}

		actual, err := c.CreateAPIClient(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleAPIClient, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid creator ID", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClient.ClientSecret = nil
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAPIClient(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateAPIClient(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			exampleInput.ClientID,
			exampleInput.ClientSecret,
			exampleInput.BelongsToUser,
		}

		db.ExpectExec(formatQueryForSQLMock(createAPIClientQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleAPIClient.CreatedOn
		}

		actual, err := c.CreateAPIClient(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_ArchiveAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClientID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, exampleAPIClientID}

		db.ExpectExec(formatQueryForSQLMock(archiveAPIClientQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleAPIClientID))

		assert.NoError(t, c.ArchiveAPIClient(ctx, exampleAPIClientID, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAPIClient(ctx, "", exampleUserID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleAPIClientID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClientID, ""))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAPIClientID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleUserID, exampleAPIClientID}

		db.ExpectExec(formatQueryForSQLMock(archiveAPIClientQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveAPIClient(ctx, exampleAPIClientID, exampleUserID))

		mock.AssertExpectationsForObjects(t, db)
	})
}
