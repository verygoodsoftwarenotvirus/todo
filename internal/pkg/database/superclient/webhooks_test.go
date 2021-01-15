package superclient

import (
	"context"
	"database/sql/driver"
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

func buildMockRowsFromWebhooks(includeCounts bool, filteredCount uint64, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := queriers.WebhooksTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, w := range webhooks {
		rowValues := []driver.Value{
			w.ID,
			w.Name,
			w.ContentType,
			w.URL,
			w.Method,
			strings.Join(w.Events, queriers.WebhooksTableEventsSeparator),
			strings.Join(w.DataTypes, queriers.WebhooksTableDataTypesSeparator),
			strings.Join(w.Topics, queriers.WebhooksTableTopicsSeparator),
			w.CreatedOn,
			w.LastUpdatedOn,
			w.ArchivedOn,
			w.BelongsToUser,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowFromWebhook(w *types.Webhook) *sqlmock.Rows {
	exampleRows := sqlmock.NewRows(queriers.WebhooksTableColumns).AddRow(
		w.ArchivedOn,
		w.BelongsToUser,
		w.Name,
		w.ContentType,
		w.URL,
		w.Method,
		strings.Join(w.Events, queriers.WebhooksTableEventsSeparator),
		strings.Join(w.DataTypes, queriers.WebhooksTableDataTypesSeparator),
		strings.Join(w.Topics, queriers.WebhooksTableTopicsSeparator),
		w.CreatedOn,
		w.LastUpdatedOn,
		w.ID,
	)

	return exampleRows
}

func TestClient_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		c, db, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("GetWebhook", mock.Anything, exampleWebhook.ID, exampleWebhook.BelongsToUser).Return(exampleWebhook, nil)

		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		mockQueryBuilder.WebhookSQLQueryBuilder.On("BuildGetWebhookQuery", exampleWebhook.ID, exampleWebhook.BelongsToUser).Return(dummyQuery, dummyArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formattedDummyQuery).
			WithArgs(interfaceToDriverValue(dummyArgs)...).
			WillReturnRows(buildMockRowsFromWebhooks(false, 0, exampleWebhook))

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		assert.NoError(t, db.ExpectationsWereMet(), "not all database query expectations were met")
		mock.AssertExpectationsForObjects(t, mockDB, mockQueryBuilder)
	})
}

func TestClient_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expected := uint64(123)

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("GetAllWebhooksCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []types.Webhook)
		exampleBucketSize := uint16(1000)

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("GetAllWebhooksCount", mock.Anything).Return(uint64(123), nil)
		mockDB.WebhookDataManager.On("GetAllWebhooks", mock.Anything, results, exampleBucketSize).Return(nil)

		err := c.GetAllWebhooks(ctx, results, exampleBucketSize)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		filter := types.DefaultQueryFilter()

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, exampleUser.ID, filter).Return(exampleWebhookList, nil)

		actual, err := c.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		filter := (*types.QueryFilter)(nil)

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("GetWebhooks", mock.Anything, exampleUser.ID, filter).Return(exampleWebhookList, nil)

		actual, err := c.GetWebhooks(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("CreateWebhook", mock.Anything, exampleInput).Return(exampleWebhook, nil)

		actual, err := c.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		var expected error

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("UpdateWebhook", mock.Anything, exampleWebhook).Return(expected)

		actual := c.UpdateWebhook(ctx, exampleWebhook)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		var expected error

		c, _, mockDB := buildTestClient(t)
		mockDB.WebhookDataManager.On("ArchiveWebhook", mock.Anything, exampleWebhook.ID, exampleWebhook.BelongsToUser).Return(expected)

		actual := c.ArchiveWebhook(ctx, exampleWebhook.ID, exampleWebhook.BelongsToUser)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
