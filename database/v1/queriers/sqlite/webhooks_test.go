package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/DATA-DOG/go-sqlmock"
	fake "github.com/brianvoe/gofakeit"
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

func buildFakeWebhook() *models.Webhook {
	return &models.Webhook{
		ID:            fake.Uint64(),
		Name:          fake.Word(),
		ContentType:   fake.MimeType(),
		URL:           fake.URL(),
		Method:        fake.HTTPMethod(),
		Events:        []string{"things"},
		DataTypes:     []string{"things"},
		Topics:        []string{"things"},
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		ArchivedOn:    nil,
		BelongsToUser: fake.Uint64(),
	}
}

func buildFakeWebhookCreationInput(webhook *models.Webhook) *models.WebhookCreationInput {
	return &models.WebhookCreationInput{
		Name:          webhook.Name,
		ContentType:   webhook.ContentType,
		URL:           webhook.URL,
		Method:        webhook.Method,
		Events:        webhook.Events,
		DataTypes:     webhook.DataTypes,
		Topics:        webhook.Topics,
		BelongsToUser: webhook.BelongsToUser,
	}
}

func TestSqlite_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhook := buildFakeWebhook()

		expectedArgCount := 2
		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = ? AND webhooks.id = ?"

		actualQuery, args := s.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleWebhook.BelongsToUser, args[0])
		assert.Equal(t, exampleWebhook.ID, args[1])
	})
}

func TestSqlite_GetWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = ? AND webhooks.id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnRows(buildMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, exampleWebhook.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, exampleWebhook.ID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID, exampleWebhook.ID).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetWebhook(ctx, exampleWebhook.ID, expectedUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetWebhookCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()
		expectedArgCount := 1
		expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? LIMIT 20"

		actualQuery, args := s.buildGetWebhookCountQuery(expectedUserID, filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, expectedUserID, args[0])
	})
}

func TestSqlite_GetWebhookCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expected := fake.Uint64()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected))

		actual, err := s.GetWebhookCount(ctx, expectedUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		expectedUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(expectedUserID).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhookCount(ctx, expectedUserID, filter)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		expected := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"

		actual := s.buildGetAllWebhooksCountQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestSqlite_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		expected := fake.Uint64()

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
		expected := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"

		actual := s.buildGetAllWebhooksQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestSqlite_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"
	expectedCountQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		exampleWebhook := buildFakeWebhook()
		expected := &models.WebhookList{
			Pagination: models.Pagination{
				Page:       1,
				TotalCount: fake.Uint64(),
			},
			Webhooks: []models.Webhook{
				*exampleWebhook,
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected.TotalCount))

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
		exampleWebhook := buildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetAllWebhooks(context.Background())
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		exampleWebhook := buildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
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

	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		expected := []models.Webhook{
			*exampleWebhook,
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
		)

		actual, err := s.GetAllWebhooksForUser(ctx, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetAllWebhooksForUser(ctx, exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetAllWebhooksForUser(ctx, exampleUserID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetAllWebhooksForUser(ctx, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleUserID := fake.Uint64()
		filter := models.DefaultQueryFilter()

		expectedArgCount := 1
		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? LIMIT 20"

		actualQuery, args := s.buildGetWebhooksQuery(exampleUserID, filter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
	})
}

func TestSqlite_GetWebhooks(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? LIMIT 20"
	expectedCountQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = ? LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhook := buildFakeWebhook()
		expected := &models.WebhookList{
			Pagination: models.Pagination{
				Page:       1,
				Limit:      20,
				TotalCount: fake.Uint64(),
			},
			Webhooks: []models.Webhook{
				*exampleWebhook,
			},
		}

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expected.TotalCount))

		actual, err := s.GetWebhooks(ctx, exampleWebhook.BelongsToUser, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := s.GetWebhooks(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleUserID := fake.Uint64()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhooks(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhook := buildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := s.GetWebhooks(ctx, exampleWebhook.BelongsToUser, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error fetching count", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhook := buildFakeWebhook()

		s, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
			buildMockRowFromWebhook(exampleWebhook),
		)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedCountQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := s.GetWebhooks(ctx, exampleWebhook.BelongsToUser, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildWebhookCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhook := buildFakeWebhook()
		expectedArgCount := 8
		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"

		actualQuery, args := s.buildWebhookCreationQuery(exampleWebhook)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleWebhook.Name, args[0])
		assert.Equal(t, exampleWebhook.ContentType, args[1])
		assert.Equal(t, exampleWebhook.URL, args[2])
		assert.Equal(t, exampleWebhook.Method, args[3])
		assert.Equal(t, strings.Join(exampleWebhook.Events, eventsSeparator), args[4])
		assert.Equal(t, strings.Join(exampleWebhook.DataTypes, typesSeparator), args[5])
		assert.Equal(t, strings.Join(exampleWebhook.Topics, topicsSeparator), args[6])
		assert.Equal(t, exampleWebhook.BelongsToUser, args[7])
	})
}

func TestSqlite_CreateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES (?,?,?,?,?,?,?,?)"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		exampleInput := buildFakeWebhookCreationInput(exampleWebhook)
		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		).WillReturnResult(exampleRows)

		expectedTimeQuery := "SELECT created_on FROM webhooks WHERE id = ?"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedTimeQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"created_on"}).AddRow(exampleWebhook.CreatedOn))

		actual, err := s.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		exampleInput := buildFakeWebhookCreationInput(exampleWebhook)

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		).WillReturnError(errors.New("blah"))

		actual, err := s.CreateWebhook(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhook := buildFakeWebhook()

		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}
		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

		actualQuery, actualArgs := s.buildUpdateWebhookQuery(exampleWebhook)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_UpdateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, updated_on = (strftime('%s','now')) WHERE belongs_to_user = ? AND id = ?"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		exampleWebhook := buildFakeWebhook()
		exampleRows := sqlmock.NewResult(int64(exampleWebhook.ID), 1)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnResult(exampleRows)

		err := s.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		s, mockDB := buildTestService(t)
		exampleWebhook := buildFakeWebhook()

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnError(errors.New("blah"))

		err := s.UpdateWebhook(ctx, exampleWebhook)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestSqlite_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s, _ := buildTestService(t)
		exampleWebhookID := fake.Uint64()
		exampleUserID := fake.Uint64()
		expectedArgCount := 2
		expectedQuery := "UPDATE webhooks SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

		actualQuery, args := s.buildArchiveWebhookQuery(exampleWebhookID, exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)

		assert.Equal(t, exampleUserID, args[0])
		assert.Equal(t, exampleWebhookID, args[1])
	})
}

func TestSqlite_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := buildFakeWebhook()
		expectedQuery := "UPDATE webhooks SET updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_user = ? AND id = ?"

		s, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
