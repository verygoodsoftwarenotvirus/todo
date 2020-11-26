package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromOAuth2Clients(includeCount bool, clients ...*types.OAuth2Client) *sqlmock.Rows {
	columns := queriers.OAuth2ClientsTableColumns

	if includeCount {
		columns = append(columns, "count")
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

		if includeCount {
			rowValues = append(rowValues, len(clients))
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
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, err := s.scanOAuth2Clients(mockRows, false)
		assert.Error(t, err)
	})

	T.Run("logs row closing errors", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return(errors.New("blah"))

		_, _, err := s.scanOAuth2Clients(mockRows, false)
		assert.NoError(t, err)
	})
}

func TestSqlite_buildGetOAuth2ClientByClientIDQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.client_id = ?"
		expectedArgs := []interface{}{
			exampleOAuth2Client.ClientID,
		}
		actualQuery, actualArgs := s.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, exampleOAuth2Client))

		actual, err := s.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous row", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientByClientIDQuery(exampleOAuth2Client.ClientID)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := s.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllOAuth2ClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL"
		actualQuery := s.buildGetAllOAuth2ClientsQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllOAuth2Clients(T *testing.T) {
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

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(
				buildMockRowsFromOAuth2Clients(
					false,
					exampleOAuth2Client,
					exampleOAuth2Client,
					exampleOAuth2Client,
				),
			)

		actual, err := s.GetAllOAuth2Clients(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllOAuth2Clients(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllOAuth2Clients(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllOAuth2ClientsQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := s.GetAllOAuth2Clients(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_GetAllOAuth2ClientsForUser(T *testing.T) {
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

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, expected...))

		actual, err := s.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, nil)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := s.GetAllOAuth2ClientsForUser(ctx, exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = ? AND oauth2_clients.id = ?"
		expectedArgs := []interface{}{
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := s.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildMockRowsFromOAuth2Clients(false, exampleOAuth2Client))

		actual, err := s.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.True(t, errors.Is(err, sql.ErrNoRows))

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(exampleOAuth2Client))

		actual, err := s.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllOAuth2ClientsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(oauth2_clients.id) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL"
		actualQuery := s.buildGetAllOAuth2ClientsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedCount := uint64(123)

		s, mockDB := buildTestService(t)
		expectedQuery := s.buildGetAllOAuth2ClientsCountQuery()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetAllOAuth2ClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetOAuth2ClientsForUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT oauth2_clients.id, oauth2_clients.name, oauth2_clients.client_id, oauth2_clients.scopes, oauth2_clients.redirect_uri, oauth2_clients.client_secret, oauth2_clients.created_on, oauth2_clients.last_updated_on, oauth2_clients.archived_on, oauth2_clients.belongs_to_user, (SELECT COUNT(*) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = ? AND items.created_on > ? AND items.created_on < ? AND items.last_updated_on > ? AND items.last_updated_on < ?) FROM oauth2_clients WHERE oauth2_clients.archived_on IS NULL AND oauth2_clients.belongs_to_user = ? AND oauth2_clients.created_on > ? AND oauth2_clients.created_on < ? AND oauth2_clients.last_updated_on > ? AND oauth2_clients.last_updated_on < ? ORDER BY oauth2_clients.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_GetOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(
				buildMockRowsFromOAuth2Clients(
					true,
					&exampleOAuth2ClientList.Clients[0],
					&exampleOAuth2ClientList.Clients[1],
					&exampleOAuth2ClientList.Clients[2],
				),
			)

		actual, err := s.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildGetOAuth2ClientsForUserQuery(exampleUser.ID, filter)

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(fakes.BuildFakeOAuth2Client()))

		actual, err := s.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "INSERT INTO oauth2_clients (name,client_id,client_secret,scopes,redirect_uri,belongs_to_user) VALUES (?,?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleOAuth2Client.Name,
			exampleOAuth2Client.ClientID,
			exampleOAuth2Client.ClientSecret,
			strings.Join(exampleOAuth2Client.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
			exampleOAuth2Client.RedirectURI,
			exampleOAuth2Client.BelongsToUser,
		}
		actualQuery, actualArgs := s.buildCreateOAuth2ClientQuery(exampleOAuth2Client)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)
		exampleRows := sqlmock.NewResult(int64(exampleOAuth2Client.ID), 1)

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildCreateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(exampleRows)

		mtt := &queriers.MockTimeTeller{}
		mtt.On("Now").Return(exampleOAuth2Client.CreatedOn)
		s.timeTeller = mtt

		actual, err := s.CreateOAuth2Client(ctx, expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mtt)
		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expectedInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildCreateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := s.CreateOAuth2Client(ctx, expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "UPDATE oauth2_clients SET client_id = ?, client_secret = ?, scopes = ?, redirect_uri = ?, last_updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleOAuth2Client.ClientID,
			exampleOAuth2Client.ClientSecret,
			strings.Join(exampleOAuth2Client.Scopes, queriers.OAuth2ClientsTableScopeSeparator),
			exampleOAuth2Client.RedirectURI,
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := s.buildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildUpdateOAuth2ClientQuery(exampleOAuth2Client)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := s.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, _ := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expectedQuery := "UPDATE oauth2_clients SET last_updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleOAuth2Client.BelongsToUser,
			exampleOAuth2Client.ID,
		}
		actualQuery, actualArgs := s.buildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildArchiveOAuth2ClientQuery(exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnError(errors.New("blah"))

		err := s.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogOAuth2ClientCreationEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntryInput := audit.BuildOAuth2ClientCreationEventEntry(exampleOAuth2Client)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogOAuth2ClientCreationEvent(ctx, exampleOAuth2Client)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_LogOAuth2ClientArchiveEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleInput := fakes.BuildFakeOAuth2Client()
		exampleAuditLogEntryInput := audit.BuildOAuth2ClientArchiveEventEntry(exampleInput.BelongsToUser, exampleInput.ID)
		exampleAuditLogEntry := converters.ConvertAuditLogEntryCreationInputToEntry(exampleAuditLogEntryInput)

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogOAuth2ClientArchiveEvent(ctx, exampleInput.BelongsToUser, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
