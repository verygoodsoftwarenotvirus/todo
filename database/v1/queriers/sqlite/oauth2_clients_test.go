package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowFromOAuth2Client(c *models.OAuth2Client) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(oauth2ClientsTableColumns).AddRow(
		c.ID,
		c.Name,
		c.ClientID,
		strings.Join(c.Scopes, scopesSeparator),
		c.RedirectURI,
		c.ClientSecret,
		c.CreatedOn,
		c.UpdatedOn,
		c.ArchivedOn,
		c.BelongsTo,
	)

	return exampleRows
}

func buildErroneousMockRowFromOAuth2Client(c *models.OAuth2Client) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(oauth2ClientsTableColumns).AddRow(
		c.ArchivedOn,
		c.Name,
		c.ClientID,
		strings.Join(c.Scopes, scopesSeparator),
		c.RedirectURI,
		c.ClientSecret,
		c.CreatedOn,
		c.UpdatedOn,
		c.BelongsTo,
		c.ID,
	)

	return exampleRows
}

func TestSqlite_buildGetOAuth2ClientByClientIDQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedClientID := "ClientID"
		expectedArgCount := 1
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND client_id = ?"

		actualQuery, args := s.buildGetOAuth2ClientByClientIDQuery(expectedClientID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedClientID, args[0].(string))
	})
}

func TestSqlite_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := "EXAMPLE"
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}

		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND client_id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleClientID).
			WillReturnRows(buildMockRowFromOAuth2Client(expected))

		actual, err := s.GetOAuth2ClientByClientID(context.Background(), exampleClientID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		exampleClientID := "EXAMPLE"
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND client_id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleClientID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetOAuth2ClientByClientID(context.Background(), exampleClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous row", func(t *testing.T) {
		exampleClientID := "EXAMPLE"
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}

		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND client_id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleClientID).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(expected))

		actual, err := s.GetOAuth2ClientByClientID(context.Background(), exampleClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllOAuth2ClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		actualQuery := s.buildGetAllOAuth2ClientsQuery()
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := []*models.OAuth2Client{
			{
				ID:        123,
				Name:      "name",
				BelongsTo: expectedUserID,
				CreatedOn: uint64(time.Now().Unix()),
			},
		}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WillReturnRows(
			buildMockRowFromOAuth2Client(expected[0]),
			buildMockRowFromOAuth2Client(expected[0]),
			buildMockRowFromOAuth2Client(expected[0]),
		)

		actual, err := s.GetAllOAuth2Clients(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllOAuth2Clients(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error executing query", func(t *testing.T) {
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllOAuth2Clients(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := []*models.OAuth2Client{
			{
				ID:        123,
				Name:      "name",
				BelongsTo: expectedUserID,
				CreatedOn: uint64(time.Now().Unix()),
			},
		}

		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(expected[0]))

		actual, err := s.GetAllOAuth2Clients(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_GetAllOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		exampleUser := &models.User{ID: 123}
		expected := []*models.OAuth2Client{
			{
				ID:        123,
				Name:      "name",
				BelongsTo: expectedUserID,
				CreatedOn: uint64(time.Now().Unix()),
			},
		}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WillReturnRows(
			buildMockRowFromOAuth2Client(expected[0]),
			buildMockRowFromOAuth2Client(expected[0]),
			buildMockRowFromOAuth2Client(expected[0]),
		)

		actual, err := s.GetAllOAuth2ClientsForUser(context.Background(), exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllOAuth2ClientsForUser(context.Background(), exampleUser.ID)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllOAuth2ClientsForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with unscannable response", func(t *testing.T) {
		expectedUserID := uint64(321)
		exampleUser := &models.User{ID: 123}
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(expected))

		actual, err := s.GetAllOAuth2ClientsForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedClientID := uint64(123)
		expectedUserID := uint64(321)
		expectedArgCount := 2
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		actualQuery, args := s.buildGetOAuth2ClientQuery(expectedClientID, expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
		assert.Equal(t, expectedClientID, args[1].(uint64))
	})
}

func TestSqlite_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
			Scopes:    []string{"things"},
		}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.BelongsTo, expected.ID).
			WillReturnRows(buildMockRowFromOAuth2Client(expected))

		actual, err := s.GetOAuth2Client(context.Background(), expected.ID, expected.BelongsTo)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
			Scopes:    []string{"things"},
		}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.BelongsTo, expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetOAuth2Client(context.Background(), expected.ID, expected.BelongsTo)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expected.BelongsTo, expected.ID).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(expected))

		actual, err := s.GetOAuth2Client(context.Background(), expected.ID, expected.BelongsTo)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetOAuth2ClientCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedUserID := uint64(321)
		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(id) FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		actualQuery, args := s.buildGetOAuth2ClientCountQuery(models.DefaultQueryFilter(), expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestSqlite_GetOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expectedQuery := "SELECT COUNT(id) FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		expectedCount := uint64(666)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetOAuth2ClientCount(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllOAuth2ClientCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expected := "SELECT COUNT(id) FROM oauth2_clients WHERE archived_on IS NULL"

		actual := s.buildGetAllOAuth2ClientCountQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestSqlite_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM oauth2_clients WHERE archived_on IS NULL"
		expectedCount := uint64(666)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actualCount, err := s.GetAllOAuth2ClientCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, actualCount)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetOAuth2ClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedUserID := uint64(321)
		expectedArgCount := 1
		expectedQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		actualQuery, args := s.buildGetOAuth2ClientsQuery(models.DefaultQueryFilter(), expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestSqlite_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedUserID := uint64(321)
		expected := &models.OAuth2ClientList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: 111,
			},
			Clients: []models.OAuth2Client{
				{
					ID:        123,
					Name:      "name",
					BelongsTo: expectedUserID,
					CreatedOn: uint64(time.Now().Unix()),
				},
			},
		}

		filter := models.DefaultQueryFilter()
		expectedListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"
		expectedCountQuery := "SELECT COUNT(id) FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromOAuth2Client(&expected.Clients[0]),
			buildMockRowFromOAuth2Client(&expected.Clients[0]),
			buildMockRowFromOAuth2Client(&expected.Clients[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected.TotalCount))

		actual, err := s.GetOAuth2Clients(ctx, filter, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with no rows returned from database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expectedListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error reading from database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expectedListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2ClientList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: 111,
			},
			Clients: []models.OAuth2Client{
				{
					ID:        123,
					Name:      "name",
					BelongsTo: expectedUserID,
					CreatedOn: uint64(time.Now().Unix()),
				},
			},
		}
		expectedListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromOAuth2Client(&expected.Clients[0]))

		actual, err := s.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2ClientList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: 0,
			},
			Clients: []models.OAuth2Client{
				{
					ID:        123,
					Name:      "name",
					BelongsTo: expectedUserID,
					CreatedOn: uint64(time.Now().Unix()),
				},
			},
		}
		expectedListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"
		expectedCountQuery := "SELECT COUNT(id) FROM oauth2_clients WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromOAuth2Client(&expected.Clients[0]),
			buildMockRowFromOAuth2Client(&expected.Clients[0]),
			buildMockRowFromOAuth2Client(&expected.Clients[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildCreateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleInput := &models.OAuth2Client{
			ClientID:     "ClientID",
			ClientSecret: "ClientSecret",
			Scopes:       []string{"blah"},
			RedirectURI:  "RedirectURI",
			BelongsTo:    123,
		}
		expectedArgCount := 6
		expectedQuery := "INSERT INTO oauth2_clients (name,client_id,client_secret,scopes,redirect_uri,belongs_to) VALUES (?,?,?,?,?,?)"

		actualQuery, args := s.buildCreateOAuth2ClientQuery(exampleInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleInput.Name, args[0].(string))
		assert.Equal(t, exampleInput.ClientID, args[1].(string))
		assert.Equal(t, exampleInput.ClientSecret, args[2].(string))
		assert.Equal(t, exampleInput.Scopes[0], args[3].(string))
		assert.Equal(t, exampleInput.RedirectURI, args[4].(string))
		assert.Equal(t, exampleInput.BelongsTo, args[5].(uint64))
	})
}

func TestSqlite_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.OAuth2ClientCreationInput{
			Name:      expected.Name,
			BelongsTo: expected.BelongsTo,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "INSERT INTO oauth2_clients (name,client_id,client_secret,scopes,redirect_uri,belongs_to) VALUES (?,?,?,?,?,?)"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ClientID,
			expected.ClientSecret,
			strings.Join(expected.Scopes, scopesSeparator),
			expected.RedirectURI,
			expected.BelongsTo,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT created_on FROM oauth2_clients WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WithArgs(expected.ID).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := s.CreateOAuth2Client(context.Background(), expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.OAuth2Client{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.OAuth2ClientCreationInput{
			Name:      expected.Name,
			BelongsTo: expected.BelongsTo,
		}
		expectedQuery := "INSERT INTO oauth2_clients (name,client_id,client_secret,scopes,redirect_uri,belongs_to) VALUES (?,?,?,?,?,?)"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ClientID,
			expected.ClientSecret,
			strings.Join(expected.Scopes, scopesSeparator),
			expected.RedirectURI,
			expected.BelongsTo,
		).WillReturnError(errors.New("blah"))

		actual, err := s.CreateOAuth2Client(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expected := &models.OAuth2Client{
			ClientID:     "ClientID",
			ClientSecret: "ClientSecret",
			Scopes:       []string{"blah"},
			RedirectURI:  "RedirectURI",
			BelongsTo:    123,
		}
		expectedArgCount := 6
		expectedQuery := "UPDATE oauth2_clients SET client_id = ?, client_secret = ?, scopes = ?, redirect_uri = ?, updated_on = (strftime('%s','now')) WHERE belongs_to = ? AND id = ?"

		actualQuery, args := s.buildUpdateOAuth2ClientQuery(expected)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expected.ClientID, args[0].(string))
		assert.Equal(t, expected.ClientSecret, args[1].(string))
		assert.Equal(t, expected.Scopes[0], args[2].(string))
		assert.Equal(t, expected.RedirectURI, args[3].(string))
		assert.Equal(t, expected.BelongsTo, args[4].(uint64))
		assert.Equal(t, expected.ID, args[5].(uint64))
	})
}

func TestSqlite_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "UPDATE oauth2_clients SET client_id = ?, client_secret = ?, scopes = ?, redirect_uri = ?, updated_on = (strftime('%s','now')) WHERE belongs_to = ? AND id = ?"
		exampleInput := &models.OAuth2Client{}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.UpdateOAuth2Client(context.Background(), exampleInput)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		expectedQuery := "UPDATE oauth2_clients SET client_id = ?, client_secret = ?, scopes = ?, redirect_uri = ?, updated_on = (strftime('%s','now')) WHERE belongs_to = ? AND id = ?"
		exampleInput := &models.OAuth2Client{}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		err := s.UpdateOAuth2Client(context.Background(), exampleInput)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveOAuth2ClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedClientID := uint64(123)
		expectedUserID := uint64(321)
		expectedArgCount := 2
		expectedQuery := "UPDATE oauth2_clients SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE belongs_to = ? AND id = ?"

		actualQuery, args := s.buildArchiveOAuth2ClientQuery(expectedClientID, expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
		assert.Equal(t, expectedClientID, args[1].(uint64))
	})
}

func TestSqlite_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "UPDATE oauth2_clients SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE belongs_to = ? AND id = ?"
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUserID, exampleClientID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error writing to database", func(t *testing.T) {
		expectedQuery := "UPDATE oauth2_clients SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE belongs_to = ? AND id = ?"
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleUserID, exampleClientID).
			WillReturnError(errors.New("blah"))

		err := s.ArchiveOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
