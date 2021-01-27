package postgres

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

func TestPostgres_ScanOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanOAuth2Clients(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, _, err := q.scanOAuth2Clients(mockRows, false)
		assert.Error(t, err)
	})
}

func TestPostgres_BuildGetOAuth2ClientByClientIDQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.client_id = $1"
		expectedArgs := []interface{}{
			exampleOAuth2Client.ClientID,
		}
		actualQuery, actualArgs := q.BuildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, 0, exampleOAuth2Client))

		actual, err := q.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous row", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := q.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildGetBatchOfOAuth2ClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.id > $1 AND oauth2_clients.id < $2"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfOAuth2ClientsQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	_q, _ := buildTestService(T)
	expectedCountQuery := _q.BuildGetAllOAuth2ClientsCountQuery()
	exampleBatchSize := uint16(1000)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.BuildGetBatchOfOAuth2ClientsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromOAuth2Clients(
					false,
					0,
					exampleOAuth2ClientList.Clients...,
				),
			)

		out := make(chan []*types.OAuth2Client)
		doneChan := make(chan bool, 1)

		err := q.GetAllOAuth2Clients(ctx, out, exampleBatchSize)
		assert.NoError(t, err)

		stillQuerying := true
		for stillQuerying {
			select {
			case batch := <-out:
				assert.NotEmpty(t, batch)
				doneChan <- true
			case <-time.After(time.Second):
				t.FailNow()
			case <-doneChan:
				stillQuerying = false
			}
		}

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching initial count", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		out := make(chan []*types.OAuth2Client)

		err := q.GetAllOAuth2Clients(ctx, out, exampleBatchSize)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.BuildGetBatchOfOAuth2ClientsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		out := make(chan []*types.OAuth2Client)

		err := q.GetAllOAuth2Clients(ctx, out, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.BuildGetBatchOfOAuth2ClientsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		out := make(chan []*types.OAuth2Client)

		err := q.GetAllOAuth2Clients(ctx, out, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedCount := uint64(20)

		begin, end := uint64(1), uint64(1001)
		expectedQuery, expectedArgs := q.BuildGetBatchOfOAuth2ClientsQuery(begin, end)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		out := make(chan []*types.OAuth2Client)

		err := q.GetAllOAuth2Clients(ctx, out, exampleBatchSize)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildGetOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = $1 AND oauth2_clients.id = $2"
		expectedArgs := []interface{}{
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := q.BuildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, 0, exampleOAuth2Client))

		actual, err := q.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := q.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildGetAllOAuth2ClientsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(oauth2_clients.id) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL"
		actualQuery := q.BuildGetAllOAuth2ClientsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedCount := uint64(123)

		q, mockDB := buildTestService(t)
		expectedQuery := q.BuildGetAllOAuth2ClientsCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := q.GetTotalOAuth2ClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildGetOAuth2ClientsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user, (SELECT COUNT(oauth2_clients.id) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = $1) as total_count, (SELECT COUNT(oauth2_clients.id) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = $2 AND oauth2_clients.created_on > $3 AND oauth2_clients.created_on < $4 AND oauth2_clients.last_updated_on > $5 AND oauth2_clients.last_updated_on < $6) as filtered_count FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = $7 AND oauth2_clients.created_on > $8 AND oauth2_clients.created_on < $9 AND oauth2_clients.last_updated_on > $10 AND oauth2_clients.last_updated_on < $11 GROUP BY oauth2_clients.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleUser.ID,
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetOAuth2ClientsQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()

		q, mockDB := buildTestService(t)
		expectedQuery, _ := q.BuildGetOAuth2ClientsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(
				buildMockRowsFromOAuth2Clients(
					true,
					exampleOAuth2ClientList.FilteredCount,
					exampleOAuth2ClientList.Clients...,
				),
			)

		actual, err := q.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, _ := q.BuildGetOAuth2ClientsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(sql.ErrNoRows)

		actual, err := q.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, _ := q.BuildGetOAuth2ClientsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := q.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		filter := types.DefaultQueryFilter()

		q, mockDB := buildTestService(t)
		expectedQuery, _ := q.BuildGetOAuth2ClientsQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := q.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildCreateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		expectedQuery := "INSERT INTO oauth2_clients (name,client_id,client_secret,scopes,redirect_uri,belongs_to_user) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleOAuth2Client.Name,
			exampleOAuth2Client.ClientID,
			exampleOAuth2Client.ClientSecret,
			strings.Join(exampleOAuth2Client.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
			exampleOAuth2Client.RedirectURI,
			exampleOAuth2Client.BelongsToUser,
		}
		actualQuery, actualArgs := q.BuildCreateOAuth2ClientQuery(expectedInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleOAuth2Client.ID, exampleOAuth2Client.CreatedOn)

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildCreateOAuth2ClientQuery(expectedInput)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		actual, err := q.CreateOAuth2Client(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildCreateOAuth2ClientQuery(expectedInput)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := q.CreateOAuth2Client(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildUpdateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "UPDATE oauth2_clients SET client_id = $1, client_secret = $2, scopes = $3, redirect_uri = $4, last_updated_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $5 AND id = $6 RETURNING last_updated_on"
		expectedArgs := []interface{}{
			exampleOAuth2Client.ClientID,
			exampleOAuth2Client.ClientSecret,
			strings.Join(exampleOAuth2Client.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
			exampleOAuth2Client.RedirectURI,
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, _ := q.BuildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"last_updated_on"}).AddRow(time.Now().Unix()))

		err := q.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, _ := q.BuildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		err := q.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_BuildArchiveOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "UPDATE oauth2_clients SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $1 AND id = $2 RETURNING archived_on"
		expectedArgs := []interface{}{
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := q.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		q, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := q.BuildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := q.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogOAuth2ClientCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntryInput := audit.BuildOAuth2ClientCreationEventEntry(exampleInput)

		expectedQuery, expectedArgs := q.BuildCreateAuditLogEntryQuery(exampleAuditLogEntryInput)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogOAuth2ClientCreationEvent(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogOAuth2ClientArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		q, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntryInput := audit.BuildOAuth2ClientArchiveEventEntry(exampleInput.BelongsToUser, exampleInput.ID)

		expectedQuery, expectedArgs := q.BuildCreateAuditLogEntryQuery(exampleAuditLogEntryInput)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		q.LogOAuth2ClientArchiveEvent(ctx, exampleInput.BelongsToUser, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
