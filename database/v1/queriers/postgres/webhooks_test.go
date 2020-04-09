package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func buildMockRowsFromWebhook(webhooks ...*models.Webhook) *sqlmock.Rows {
	includeCount := len(webhooks) > 1
	columns := webhooksTableColumns

	if includeCount {
		columns = append(columns, "count")
	}
	exampleRows := sqlmock.NewRows(columns)

	for _, w := range webhooks {
		rowValues := []driver.Value{
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
		}

		if includeCount {
			rowValues = append(rowValues, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

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

func TestPostgres_buildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = $1 AND webhooks.id = $2"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}

		actualQuery, actualArgs := p.buildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.belongs_to_user = $1 AND webhooks.id = $2"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnRows(buildMockRowsFromWebhook(exampleWebhook))

		actual, err := p.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(exampleWebhook.BelongsToUser, exampleWebhook.ID).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := p.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expected := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"

		actual := p.buildGetAllWebhooksCountQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestPostgres_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleCount := uint64(123)

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(exampleCount))

		actual, err := p.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAllWebhooksCount(ctx)
		assert.Error(t, err)
		assert.Zero(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetAllWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expected := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"

		actual := p.buildGetAllWebhooksQuery()
		assert.Equal(t, expected, actual)
	})
}

func TestPostgres_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user FROM webhooks WHERE webhooks.archived_on IS NULL"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()
		exampleWebhookList.Limit = 0

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).WillReturnRows(
			buildMockRowsFromWebhook(
				&exampleWebhookList.Webhooks[0],
				&exampleWebhookList.Webhooks[1],
				&exampleWebhookList.Webhooks[2],
			),
		)

		actual, err := p.GetAllWebhooks(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := p.GetAllWebhooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUser := fakemodels.BuildFakeUser()
		filter := fakemodels.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user, COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = $1 AND webhooks.created_on > $2 AND webhooks.created_on < $3 AND webhooks.updated_on > $4 AND webhooks.updated_on < $5 GROUP BY webhooks.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleUser.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}

		actualQuery, actualArgs := p.buildGetWebhooksQuery(exampleUser.ID, filter)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	expectedListQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.updated_on, webhooks.archived_on, webhooks.belongs_to_user, COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_user = $1 GROUP BY webhooks.id LIMIT 20"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowsFromWebhook(
					&exampleWebhookList.Webhooks[0],
					&exampleWebhookList.Webhooks[1],
					&exampleWebhookList.Webhooks[2],
				),
			)

		actual, err := p.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("surfaces sql.ErrNoRows", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(sql.ErrNoRows)

		actual, err := p.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, sql.ErrNoRows, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error querying database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnError(errors.New("blah"))

		actual, err := p.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		ctx := context.Background()
		filter := models.DefaultQueryFilter()
		exampleWebhook := fakemodels.BuildFakeWebhook()

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedListQuery)).
			WillReturnRows(buildErroneousMockRowFromWebhook(exampleWebhook))

		actual, err := p.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildWebhookCreationQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_on"
		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		}

		actualQuery, actualArgs := p.buildWebhookCreationQuery(exampleWebhook)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_CreateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "INSERT INTO webhooks (name,content_type,url,method,events,data_types,topics,belongs_to_user) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_on"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleWebhook.ID, exampleWebhook.CreatedOn)

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		).WillReturnRows(exampleRows)

		actual, err := p.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error interacting with database", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		p, mockDB := buildTestService(t)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
		).WillReturnError(errors.New("blah"))

		actual, err := p.CreateWebhook(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET name = $1, content_type = $2, url = $3, method = $4, events = $5, data_types = $6, topics = $7, updated_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $8 AND id = $9 RETURNING updated_on"
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

		actualQuery, actualArgs := p.buildUpdateWebhookQuery(exampleWebhook)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_UpdateWebhook(T *testing.T) {
	T.Parallel()

	expectedQuery := "UPDATE webhooks SET name = $1, content_type = $2, url = $3, method = $4, events = $5, data_types = $6, topics = $7, updated_on = extract(epoch FROM NOW()) WHERE belongs_to_user = $8 AND id = $9 RETURNING updated_on"

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		p, mockDB := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		exampleRows := sqlmock.NewRows([]string{"updated_on"}).AddRow(uint64(time.Now().Unix()))
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, eventsSeparator),
			strings.Join(exampleWebhook.DataTypes, typesSeparator),
			strings.Join(exampleWebhook.Topics, topicsSeparator),
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnRows(exampleRows)

		err := p.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})

	T.Run("with error from database", func(t *testing.T) {
		ctx := context.Background()
		p, mockDB := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).WithArgs(
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

		err := p.UpdateWebhook(ctx, exampleWebhook)
		assert.Error(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}

func TestPostgres_buildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleWebhook := fakemodels.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2 RETURNING archived_on"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		}

		actualQuery, actualArgs := p.buildArchiveWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToUser)
		ensureArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleWebhook := fakemodels.BuildFakeWebhook()
		expectedQuery := "UPDATE webhooks SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2 RETURNING archived_on"

		p, mockDB := buildTestService(t)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).WithArgs(
			exampleWebhook.BelongsToUser,
			exampleWebhook.ID,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
