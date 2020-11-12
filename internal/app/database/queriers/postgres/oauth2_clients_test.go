package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromOAuth2Client(clients ...*types.OAuth2Client) *sqlmock.Rows {
	columns := oauth2ClientsTableColumns
	exampleRows := sqlmock.NewRows(columns)

	for _, c := range clients {
		rowValues := []driver.Value{
			c.ID,
			c.Name,
			c.ClientID,
			strings.Join(c.Scopes, scopesSeparator),
			c.RedirectURI,
			c.ClientSecret,
			c.CreatedOn,
			c.LastUpdatedOn,
			c.ArchivedOn,
			c.BelongsToUser,
		}
		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromOAuth2Client(c *types.OAuth2Client) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(oauth2ClientsTableColumns).AddRow(
		c.ArchivedOn,
		c.Name,
		c.ClientID,
		strings.Join(c.Scopes, scopesSeparator),
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
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, err := p.scanOAuth2Clients(mockRows)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, err := p.scanOAuth2Clients(mockRows)
		assert.NoError(t, err)
	})
}

func TestPostgres_buildGetOAuth2ClientByClientIDQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.client_id = $1"
		expectedArgs := []interface{}{
			exampleOAuth2Client.ClientID,
		}
		actualQuery, actualArgs := p.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous row", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllOAuth2ClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL"
		actualQuery := p.buildGetAllOAuth2ClientsQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expected := []*types.OAuth2Client{
			exampleOAuth2Client,
			exampleOAuth2Client,
			exampleOAuth2Client,
		}

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(
				buildMockRowsFromOAuth2Client(
					exampleOAuth2Client,
					exampleOAuth2Client,
					exampleOAuth2Client,
				),
			)

		actual, err := p.GetAllOAuth2Clients(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetAllOAuth2Clients(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAllOAuth2Clients(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetAllOAuth2Clients(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_GetAllOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()

		expected := []*types.OAuth2Client{
			&exampleOAuth2ClientList.Clients[0],
			&exampleOAuth2ClientList.Clients[1],
			&exampleOAuth2ClientList.Clients[2],
		}

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Client(expected...))

		actual, err := p.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = $1 AND oauth2_clients.id = $2"
		expectedArgs := []interface{}{
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := p.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllOAuth2ClientsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(oauth2_clients.id) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL"
		actualQuery := p.buildGetAllOAuth2ClientsCountQuery()

		ensureArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedCount := uint64(123)

		p, mockDB := buildTestService(t)
		expectedQuery := p.buildGetAllOAuth2ClientsCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := p.GetAllOAuth2ClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetOAuth2ClientsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = $1 AND oauth2_clients.created_on > $2 AND oauth2_clients.created_on < $3 AND oauth2_clients.last_updated_on > $4 AND oauth2_clients.last_updated_on < $5 ORDER BY oauth2_clients.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		p, mockDB := buildTestService(t)
		expectedQuery, _ := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(
				buildMockRowsFromOAuth2Client(
					&exampleOAuth2ClientList.Clients[0],
					&exampleOAuth2ClientList.Clients[1],
					&exampleOAuth2ClientList.Clients[2],
				),
			)

		actual, err := p.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		expectedQuery, _ := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		expectedQuery, _ := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := p.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		filter := types.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		expectedQuery, _ := p.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := p.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildCreateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "INSERT INTO oauth2_clients (name,client_id,client_secret,scopes,redirect_uri,belongs_to_user) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleOAuth2Client.Name,
			exampleOAuth2Client.ClientID,
			exampleOAuth2Client.ClientSecret,
			strings.Join(exampleOAuth2Client.Scopes, scopesSeparator),
			exampleOAuth2Client.RedirectURI,
			exampleOAuth2Client.BelongsToUser,
		}
		actualQuery, actualArgs := p.buildCreateOAuth2ClientQuery(exampleOAuth2Client)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildCreateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		actual, err := p.CreateOAuth2Client(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildCreateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := p.CreateOAuth2Client(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "UPDATE oauth2_clients SET client_id = $1, client_secret = $2, scopes = $3, redirect_uri = $4, last_updated_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $5 AND id = $6 RETURNING last_updated_on"
		expectedArgs := []interface{}{
			exampleOAuth2Client.ClientID,
			exampleOAuth2Client.ClientSecret,
			strings.Join(exampleOAuth2Client.Scopes, scopesSeparator),
			exampleOAuth2Client.RedirectURI,
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := p.buildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		p, mockDB := buildTestService(t)
		expectedQuery, _ := p.buildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"last_updated_on"}).AddRow(time.Now().Unix()))

		err := p.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, _ := p.buildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		err := p.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		p, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "UPDATE oauth2_clients SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $1 AND id = $2 RETURNING archived_on"
		expectedArgs := []interface{}{
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := p.buildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
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

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		p, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := p.buildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := p.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogOAuth2ClientCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntryInput := audit.BuildOAuth2ClientCreationEventEntry(exampleInput)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogOAuth2ClientCreationEvent(ctx, exampleInput)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_LogOAuth2ClientArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntryInput := audit.BuildOAuth2ClientArchiveEventEntry(exampleInput.BelongsToUser, exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogOAuth2ClientArchiveEvent(ctx, exampleInput.BelongsToUser, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
