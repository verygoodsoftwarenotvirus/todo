package mariadb

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
		w.BelongsTo,
	)

	return exampleRows
}

func buildErroneousMockRowFromWebhook(w *models.Webhook) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(webhooksTableColumns).AddRow(
		w.ArchivedOn,
		w.BelongsTo,
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

func TestMariaDB_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleWebhookID := uint64(123)
		exampleUserID := uint64(321)

		expectedArgCount := 2
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE belongs_to = ? AND id = ?"

		actualQuery, args := m.buildGetWebhookQuery(exampleWebhookID, exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
		assert.Equal(t, exampleWebhookID, args[1].(uint64))
	})
}

func TestMariaDB_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE belongs_to = ? AND id = ?"
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			Events:    []string{"things"},
			DataTypes: []string{"things"},
			Topics:    []string{"things"},
		}
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildMockRowFromWebhook(expected))

		actual, err := m.GetWebhook(context.Background(), expected.ID, expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE belongs_to = ? AND id = ?"
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			Events:    []string{"things"},
			DataTypes: []string{"things"},
			Topics:    []string{"things"},
		}
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetWebhook(context.Background(), expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE belongs_to = ? AND id = ?"
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetWebhook(context.Background(), expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		ctx := context.Background()
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE belongs_to = ? AND id = ?"
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			Events:    []string{"things"},
			DataTypes: []string{"things"},
			Topics:    []string{"things"},
		}
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, expected.ID).
			WillReturnRows(buildErroneousMockRowFromWebhook(expected))

		actual, err := m.GetWebhook(ctx, expected.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetWebhookCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expectedUserID := uint64(123)
		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		actualQuery, args := m.buildGetWebhookCountQuery(models.DefaultQueryFilter(), expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestMariaDB_GetWebhookCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		expected := uint64(321)
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := m.GetWebhookCount(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"
		expectedUserID := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetWebhookCount(context.Background(), models.DefaultQueryFilter(), expectedUserID)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"

		actual := m.buildGetAllWebhooksCountQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestMariaDB_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"
		expected := uint64(321)

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := m.GetAllWebhooksCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		expectedQuery := "SELECT COUNT(id) FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllWebhooksCount(context.Background())
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetAllWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		expected := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		actual := m.buildGetAllWebhooksQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestMariaDB_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedCount := uint64(321)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
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

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := m.GetAllWebhooks(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
		example := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(example))

		actual, err := m.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		expectedCount := uint64(321)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
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

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_GetAllWebhooksForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
		expected := []models.Webhook{
			{
				ID:   123,
				Name: "name",
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected[0]),
			buildMockRowFromWebhook(&expected[0]),
			buildMockRowFromWebhook(&expected[0]),
		)

		actual, err := m.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		exampleUser := &models.User{ID: 123}
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
		expected := []models.Webhook{
			{
				ID:   123,
				Name: "name",
			},
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(&expected[0]))

		actual, err := m.GetAllWebhooksForUser(context.Background(), exampleUser.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(123)
		m, _ := buildTestService(t)

		expectedArgCount := 1
		expectedQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL AND belongs_to = ? LIMIT 20"

		actualQuery, args := m.buildGetWebhooksQuery(models.DefaultQueryFilter(), exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}

func TestMariaDB_GetWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedCount := uint64(321)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
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

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		actual, err := m.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := m.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(expected))

		actual, err := m.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		exampleUserID := uint64(123)
		expectedCount := uint64(321)
		expectedListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
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

		m, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
			buildMockRowFromWebhook(&expected.Webhooks[0]),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := m.GetWebhooks(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildWebhookCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleInput := &models.Webhook{
			Name:        "name",
			ContentType: "application/json",
			URL:         "https://verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{},
			DataTypes:   []string{},
			Topics:      []string{},
			BelongsTo:   1,
		}
		expectedArgCount := 8
		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to,created_on) VALUES (?,?,?,?,?,?,?,?,UNIX_TIMESTAMP())"

		actualQuery, args := m.buildWebhookCreationQuery(exampleInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.WebhookCreationInput{
			Name:      expected.Name,
			BelongsTo: expected.BelongsTo,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to,created_on) VALUES (?,?,?,?,?,?,?,?,UNIX_TIMESTAMP())"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsTo,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT created_on FROM webhooks WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(expected.CreatedOn))

		actual, err := m.CreateWebhook(context.Background(), expectedInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		expectedUserID := uint64(321)
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			BelongsTo: expectedUserID,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedInput := &models.WebhookCreationInput{
			Name:      expected.Name,
			BelongsTo: expected.BelongsTo,
		}
		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to,created_on) VALUES (?,?,?,?,?,?,?,?,UNIX_TIMESTAMP())"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsTo,
		).WillReturnError(errors.New("blah"))

		actual, err := m.CreateWebhook(context.Background(), expectedInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleInput := &models.Webhook{
			Name:        "name",
			ContentType: "application/json",
			URL:         "https://verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{},
			DataTypes:   []string{},
			Topics:      []string{},
			BelongsTo:   1,
		}
		expectedArgCount := 9
		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to = ? AND id = ?"

		actualQuery, args := m.buildUpdateWebhookQuery(exampleInput)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestMariaDB_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, mockDB := buildTestService(t)
		expected := &models.Webhook{
			Name:        "name",
			ContentType: "application/json",
			URL:         "https://verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{},
			DataTypes:   []string{},
			Topics:      []string{},
			BelongsTo:   1,
		}
		exampleRows := sqlmock.NewResult(int64(expected.ID), 1)
		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to = ? AND id = ?"

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsTo,
			expected.ID,
		).WillReturnResult(exampleRows)

		err := m.UpdateWebhook(context.Background(), expected)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		m, mockDB := buildTestService(t)
		expected := &models.Webhook{
			Name:        "name",
			ContentType: "application/json",
			URL:         "https://verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{},
			DataTypes:   []string{},
			Topics:      []string{},
			BelongsTo:   1,
		}
		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = UNIX_TIMESTAMP() WHERE belongs_to = ? AND id = ?"

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.Name,
			expected.ContentType,
			expected.URL,
			expected.Method,
			strings.Join(expected.Events, eventsSeparator),
			strings.Join(expected.DataTypes, typesSeparator),
			strings.Join(expected.Topics, topicsSeparator),
			expected.BelongsTo,
			expected.ID,
		).WillReturnError(errors.New("blah"))

		err := m.UpdateWebhook(context.Background(), expected)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestMariaDB_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		m, _ := buildTestService(t)
		exampleWebhookID := uint64(123)
		exampleUserID := uint64(321)
		expectedArgCount := 2
		expectedQuery := "UPDATE webhooks SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		actualQuery, args := m.buildArchiveWebhookQuery(exampleWebhookID, exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
		assert.Equal(t, exampleWebhookID, args[1].(uint64))
	})
}

func TestMariaDB_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Webhook{
			ID:        123,
			Name:      "name",
			BelongsTo: 321,
			CreatedOn: uint64(time.Now().Unix()),
		}
		expectedQuery := "UPDATE webhooks SET updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to = ? AND id = ?"

		m, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			expected.BelongsTo,
			expected.ID,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := m.ArchiveWebhook(context.Background(), expected.ID, expected.BelongsTo)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
