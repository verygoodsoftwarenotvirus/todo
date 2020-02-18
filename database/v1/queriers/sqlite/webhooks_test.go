package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowFromWebhook(w *models.Webhook) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(webhooksTableColumns).AddRow(
		w.ID,
		w.Name,
		w.ContentType,
		w.URL,
		w.Method,
		strings.Join(w.Events, eventsSeparator),
		strings.Join(w.DataTypes, typesSeparator),
		strings.Join(w.Topics, topicsSeparator),
		w.CreatedOn,
		w.UpdatedOn,
		w.ArchivedOn,
		w.BelongsToUser,
	)

	return exampleRows
}

func buildErroneousMockRowFromWebhook(w *models.Webhook) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(webhooksTableColumns).AddRow(
		w.ArchivedOn,
		w.BelongsToUser,
		w.Name,
		w.ContentType,
		w.URL,
		w.Method,
		strings.Join(w.Events, eventsSeparator),
		strings.Join(w.DataTypes, typesSeparator),
		strings.Join(w.Topics, topicsSeparator),
		w.CreatedOn,
		w.UpdatedOn,
		w.ID,
	)

	return exampleRows
}

func TestSqlite_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhookID := uint64(123)
		exampleUserID := uint64(321)

		expectedArgCount := 2
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE belongs_to_user = ? AND id = ?"

		actualQuery, args := s.buildGetWebhookQuery(exampleWebhookID, exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
		assert.Equal(t, exampleWebhookID, args[1].(uint64))
	})
}

func TestSqlite_GetWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			Events:    []string{"things"},
			DataTypes: []string{"things"},
			Topics:    []string{"things"},
		}
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildMockRowFromWebhook(expected))

		actual, err := s.GetWebhook(context.Background(), expected.ID, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			Events:    []string{"things"},
			DataTypes: []string{"things"},
			Topics:    []string{"things"},
		}
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhook(context.Background(), expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhook(context.Background(), expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			Events:    []string{"things"},
			DataTypes: []string{"things"},
			Topics:    []string{"things"},
		}
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildErroneousMockRowFromWebhook(expected))

		actual, err := s.GetWebhook(ctx, expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetWebhookCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expectedUserID := uint64(123)
		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL AND belongs_to_user = ? LIMIT 20"

		actualQuery, args := s.buildGetWebhookCountQuery(models.DefaultQueryFilter(), expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestSqlite_GetWebhookCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL AND belongs_to_user = ? LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(321)
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := s.GetWebhookCount(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		expectedUserID := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhookCount(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expected := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"

		actual := s.buildGetAllWebhooksCountQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestSqlite_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(321)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := s.GetAllWebhooksCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooksCount(context.Background())
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expected := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE archived_on IS NULL"

		actual := s.buildGetAllWebhooksQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestSqlite_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		expectedCount := uint64(321)
		expectedCountQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"
		expected := &models.WebhookList{
			Pagination: models.Pagination{
				Page:       1,
				TotalCount: expectedCount,
			},
			Webhooks: []models.Webhook{
				{
					ID:   123,
					Name: "name",
				},
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := s.GetAllWebhooks(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		example := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(example))

		actual, err := s.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		expectedCount := uint64(321)
		expectedCountQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"
		expected := &models.WebhookList{
			Pagination: models.Pagination{
				TotalCount: expectedCount,
			},
			Webhooks: []models.Webhook{
				{
					ID:   123,
					Name: "name",
				},
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_GetAllWebhooksForUser(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expected := []models.Webhook{
			{
				ID:   123,
				Name: "name",
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected[0]),
			buildMockRowFromWebhook(&expected[0]),
			buildMockRowFromWebhook(&expected[0]),
		)

		actual, err := s.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expected := []models.Webhook{
			{
				ID:   123,
				Name: "name",
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(&expected[0]))

		actual, err := s.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(123)
		s, _ := buildTestService(t)

		expectedArgCount := 1
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE archived_on IS NULL AND belongs_to_user = ? LIMIT 20"

		actualQuery, args := s.buildGetWebhooksQuery(models.DefaultQueryFilter(), exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestSqlite_GetWebhooks(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to_user FROM webhooks WHERE archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedCount := uint64(321)
		expectedCountQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"
		expected := &models.WebhookList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Webhooks: []models.Webhook{
				{
					ID:   123,
					Name: "name",
				},
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := s.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		exampleUserID := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		exampleUserID := uint64(123)

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		exampleUserID := uint64(123)
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(expected))

		actual, err := s.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedCount := uint64(321)
		expectedCountQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"
		expected := &models.WebhookList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: expectedCount,
			},
			Webhooks: []models.Webhook{
				{
					ID:   123,
					Name: "name",
				},
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildWebhookCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleInput := &models.Webhook{
			Name:          "name",
			ContentType:   "application/json",
			URL:           "https://verygoodsoftwarenotvirus.ru",
			Method:        http.MethodPatch,
			Events:        []string{},
			DataTypes:     []string{},
			Topics:        []string{},
			BelongsToUser: 1,
		}
		expectedArgCount := 8
		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"

		actualQuery, args := s.buildWebhookCreationQuery(exampleInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_CreateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Webhook{
			ID:            123,
			Name:          "name",
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedInput := &models.WebhookCreationInput{
			Name:          expected.Name,
			BelongsToUser: expected.BelongsToUser,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsToUser,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT created_on FROM webhooks WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := s.CreateWebhook(context.Background(), expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Webhook{
			ID:            123,
			Name:          "name",
			BelongsToUser: expectedUserID,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedInput := &models.WebhookCreationInput{
			Name:          expected.Name,
			BelongsToUser: expected.BelongsToUser,
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsToUser,
		).WillReturnError(errors.New("blah"))

		actual, err := s.CreateWebhook(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleInput := &models.Webhook{
			Name:          "name",
			ContentType:   "application/json",
			URL:           "https://verygoodsoftwarenotvirus.ru",
			Method:        http.MethodPatch,
			Events:        []string{},
			DataTypes:     []string{},
			Topics:        []string{},
			BelongsToUser: 1,
		}
		expectedArgCount := 9
		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

		actualQuery, args := s.buildUpdateWebhookQuery(exampleInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestSqlite_UpdateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		s, mockDB := buildTestService(t)
		expected := &models.Webhook{
			Name:          "name",
			ContentType:   "application/json",
			URL:           "https://verygoodsoftwarenotvirus.ru",
			Method:        http.MethodPatch,
			Events:        []string{},
			DataTypes:     []string{},
			Topics:        []string{},
			BelongsToUser: 1,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsToUser,
			expected.ID,
		).WillReturnResult(exampleRows)

		err := s.UpdateWebhook(context.Background(), expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		s, mockDB := buildTestService(t)
		expected := &models.Webhook{
			Name:          "name",
			ContentType:   "application/json",
			URL:           "https://verygoodsoftwarenotvirus.ru",
			Method:        http.MethodPatch,
			Events:        []string{},
			DataTypes:     []string{},
			Topics:        []string{},
			BelongsToUser: 1,
		}

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsToUser,
			expected.ID,
		).WillReturnError(errors.New("blah"))

		err := s.UpdateWebhook(context.Background(), expected)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhookID := uint64(123)
		exampleUserID := uint64(321)
		expectedArgCount := 2
		expectedQuery := "UPDATE webhooks SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

		actualQuery, args := s.buildArchiveWebhookQuery(exampleWebhookID, exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
		assert.Equal(t, exampleWebhookID, args[1].(uint64))
	})
}

func TestSqlite_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Webhook{
			ID:            123,
			Name:          "name",
			BelongsToUser: 321,
			CreatedOn:     uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE webhooks SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.BelongsToUser,
			expected.ID,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveWebhook(context.Background(), expected.ID, expected.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
