package base

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBase_BuildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "SELECT webhooks.id, webhooks.external_id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ? AND webhooks.id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToAccount,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.BuildGetWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestBase_BuildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := q.BuildGetAllWebhooksCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestBase_BuildGetBatchOfWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT webhooks.id, webhooks.external_id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account FROM webhooks WHERE webhooks.id > ? AND webhooks.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfWebhooksQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestBase_BuildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT webhooks.id, webhooks.external_id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account, (SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ?) as total_count, (SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.last_updated_on > ? AND webhooks.last_updated_on < ?) as filtered_count FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.last_updated_on > ? AND webhooks.last_updated_on < ? GROUP BY webhooks.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetWebhooksQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestBase_BuildCreateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleWebhook.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO webhooks (external_id,name,content_type,url,method,events,data_types,topics,belongs_to_account) VALUES (?,?,?,?,?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleWebhook.ExternalID,
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(exampleWebhook.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(exampleWebhook.Topics, querybuilding.WebhooksTableTopicsSeparator),
			exampleWebhook.BelongsToAccount,
		}
		actualQuery, actualArgs := q.BuildCreateWebhookQuery(exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestBase_BuildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET name = ?, content_type = ?, url = ?, method = ?, events = ?, data_types = ?, topics = ?, last_updated_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(exampleWebhook.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(exampleWebhook.Topics, querybuilding.WebhooksTableTopicsSeparator),
			exampleWebhook.BelongsToAccount,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateWebhookQuery(exampleWebhook)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestBase_BuildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET last_updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToAccount,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveWebhookQuery(exampleWebhook.ID, exampleWebhook.BelongsToAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
