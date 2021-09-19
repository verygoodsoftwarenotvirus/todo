package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildMockRowsFromWebhooks(includeCounts bool, filteredCount uint64, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := webhooksTableColumns

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
			strings.Join(w.Events, webhooksTableEventsSeparator),
			strings.Join(w.DataTypes, webhooksTableDataTypesSeparator),
			strings.Join(w.Topics, webhooksTableTopicsSeparator),
			w.CreatedOn,
			w.LastUpdatedOn,
			w.ArchivedOn,
			w.BelongsToAccount,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func buildErroneousMockRowsFromWebhooks(includeCounts bool, filteredCount uint64, webhooks ...*types.Webhook) *sqlmock.Rows {
	columns := webhooksTableColumns

	if includeCounts {
		columns = append(columns, "filtered_count", "total_count")
	}

	exampleRows := sqlmock.NewRows(columns)

	for _, w := range webhooks {
		rowValues := []driver.Value{
			w.ArchivedOn,
			strings.Join(w.Events, webhooksTableEventsSeparator),
			strings.Join(w.DataTypes, webhooksTableDataTypesSeparator),
			strings.Join(w.Topics, webhooksTableTopicsSeparator),
			w.ID,
			w.Name,
			w.ContentType,
			w.URL,
			w.Method,
			w.CreatedOn,
			w.LastUpdatedOn,
			w.BelongsToAccount,
		}

		if includeCounts {
			rowValues = append(rowValues, filteredCount, len(webhooks))
		}

		exampleRows.AddRow(rowValues...)
	}

	return exampleRows
}

func TestQuerier_ScanWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("surfaces row errs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		q, _ := buildTestClient(t)
		mockRows := &database.MockResultIterator{}

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("blah"))

		_, _, _, err := q.scanWebhooks(ctx, mockRows, false)
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

		_, _, _, err := q.scanWebhooks(ctx, mockRows, false)
		assert.Error(t, err)
	})
}

func TestQuerier_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleAccountID, exampleWebhook.ID}

		db.ExpectQuery(formatQueryForSQLMock(getWebhookQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromWebhooks(false, 0, exampleWebhook))

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleAccountID)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetWebhook(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid response from database", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleAccountID, exampleWebhook.ID}

		db.ExpectQuery(formatQueryForSQLMock(getWebhookQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRowsFromWebhooks(false, 0, exampleWebhook))

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetAllWebhooksCount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := uint64(123)
		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectQuery(formatQueryForSQLMock(getAllWebhooksCountQuery)).
			WithArgs().
			WillReturnRows(newCountDBRowResponse(expected))

		actual, err := c.GetAllWebhooksCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_GetWebhooks(T *testing.T) {
	T.Parallel()

	exampleAccountID := fakes.BuildFakeID()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"webhooks",
			nil,
			nil,
			"belongs_to_account",
			webhooksTableColumns,
			exampleAccountID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromWebhooks(
				true,
				exampleWebhookList.FilteredCount,
				exampleWebhookList.Webhooks...,
			))

		actual, err := c.GetWebhooks(ctx, exampleAccountID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		exampleWebhookList := fakes.BuildFakeWebhookList()
		exampleWebhookList.Page = 0
		exampleWebhookList.Limit = 0
		filter := (*types.QueryFilter)(nil)

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"webhooks",
			nil,
			nil,
			"belongs_to_account",
			webhooksTableColumns,
			exampleAccountID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildMockRowsFromWebhooks(
				true,
				exampleWebhookList.FilteredCount,
				exampleWebhookList.Webhooks...,
			))

		actual, err := c.GetWebhooks(ctx, exampleAccountID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhookList, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		filter := types.DefaultQueryFilter()
		c, _ := buildTestClient(t)

		actual, err := c.GetWebhooks(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"webhooks",
			nil,
			nil,
			"belongs_to_account",
			webhooksTableColumns,
			exampleAccountID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetWebhooks(ctx, exampleAccountID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with erroneous database response", func(t *testing.T) {
		t.Parallel()

		filter := types.DefaultQueryFilter()

		ctx := context.Background()
		c, db := buildTestClient(t)

		query, args := c.buildListQuery(
			ctx,
			"webhooks",
			nil,
			nil,
			"belongs_to_account",
			webhooksTableColumns,
			exampleAccountID,
			false,
			filter,
		)

		db.ExpectQuery(formatQueryForSQLMock(query)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetWebhooks(ctx, exampleAccountID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookDatabaseCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			exampleInput.ContentType,
			exampleInput.URL,
			exampleInput.Method,
			strings.Join(exampleInput.Events, webhooksTableEventsSeparator),
			strings.Join(exampleInput.DataTypes, webhooksTableDataTypesSeparator),
			strings.Join(exampleInput.Topics, webhooksTableTopicsSeparator),
			exampleInput.BelongsToAccount,
		}

		db.ExpectExec(formatQueryForSQLMock(createWebhookQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhook.ID))

		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleWebhook, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		actual, err := c.CreateWebhook(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error executing creation query", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookDatabaseCreationInputFromWebhook(exampleWebhook)

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{
			exampleInput.ID,
			exampleInput.Name,
			exampleInput.ContentType,
			exampleInput.URL,
			exampleInput.Method,
			strings.Join(exampleInput.Events, webhooksTableEventsSeparator),
			strings.Join(exampleInput.DataTypes, webhooksTableDataTypesSeparator),
			strings.Join(exampleInput.Topics, webhooksTableTopicsSeparator),
			exampleInput.BelongsToAccount,
		}

		db.ExpectExec(formatQueryForSQLMock(createWebhookQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		c.timeFunc = func() uint64 {
			return exampleWebhook.CreatedOn
		}

		actual, err := c.CreateWebhook(ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestQuerier_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleWebhookID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleAccountID, exampleWebhookID}

		db.ExpectExec(formatQueryForSQLMock(archiveWebhookQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleWebhookID))

		actual := c.ArchiveWebhook(ctx, exampleWebhookID, exampleAccountID)
		assert.NoError(t, actual)

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveWebhook(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fakes.BuildFakeID()

		ctx := context.Background()
		c, _ := buildTestClient(t)

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhookID, ""))
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		exampleWebhookID := fakes.BuildFakeID()

		ctx := context.Background()
		c, db := buildTestClient(t)

		args := []interface{}{exampleAccountID, exampleWebhookID}

		db.ExpectExec(formatQueryForSQLMock(archiveWebhookQuery)).
			WithArgs(interfaceToDriverValue(args)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhookID, exampleAccountID))

		mock.AssertExpectationsForObjects(t, db)
	})
}
