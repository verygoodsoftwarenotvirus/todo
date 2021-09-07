package sqlite

import (
	"context"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_BuildGetWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ? AND webhooks.id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToAccount,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.BuildGetWebhookQuery(ctx, exampleWebhook.ID, exampleWebhook.BelongsToAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAllWebhooksCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		expectedQuery := "SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL"
		actualQuery := q.BuildGetAllWebhooksCountQuery(ctx)

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestSqlite_BuildGetBatchOfWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account FROM webhooks WHERE webhooks.id > ? AND webhooks.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfWebhooksQuery(ctx, beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetWebhooksQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT webhooks.id, webhooks.name, webhooks.content_type, webhooks.url, webhooks.method, webhooks.events, webhooks.data_types, webhooks.topics, webhooks.created_on, webhooks.last_updated_on, webhooks.archived_on, webhooks.belongs_to_account, (SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ?) as total_count, (SELECT COUNT(webhooks.id) FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.last_updated_on > ? AND webhooks.last_updated_on < ?) as filtered_count FROM webhooks WHERE webhooks.archived_on IS NULL AND webhooks.belongs_to_account = ? AND webhooks.created_on > ? AND webhooks.created_on < ? AND webhooks.last_updated_on > ? AND webhooks.last_updated_on < ? GROUP BY webhooks.id LIMIT 20 OFFSET 180"
		expectedArgs := []interface{}{
			exampleAccount.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
			exampleAccount.ID,
			exampleAccount.ID,
			filter.CreatedAfter,
			filter.CreatedBefore,
			filter.UpdatedAfter,
			filter.UpdatedBefore,
		}
		actualQuery, actualArgs := q.BuildGetWebhooksQuery(ctx, exampleAccount.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildCreateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		expectedQuery := "INSERT INTO webhooks (id,name,content_type,url,method,events,data_types,topics,belongs_to_account) VALUES (?,?,?,?,?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleWebhook.ID,
			exampleWebhook.Name,
			exampleWebhook.ContentType,
			exampleWebhook.URL,
			exampleWebhook.Method,
			strings.Join(exampleWebhook.Events, querybuilding.WebhooksTableEventsSeparator),
			strings.Join(exampleWebhook.DataTypes, querybuilding.WebhooksTableDataTypesSeparator),
			strings.Join(exampleWebhook.Topics, querybuilding.WebhooksTableTopicsSeparator),
			exampleWebhook.BelongsToAccount,
		}
		actualQuery, actualArgs := q.BuildCreateWebhookQuery(ctx, exampleInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildUpdateWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

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
		actualQuery, actualArgs := q.BuildUpdateWebhookQuery(ctx, exampleWebhook)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildArchiveWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "UPDATE webhooks SET last_updated_on = (strftime('%s','now')), archived_on = (strftime('%s','now')) WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleWebhook.BelongsToAccount,
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveWebhookQuery(ctx, exampleWebhook.ID, exampleWebhook.BelongsToAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestSqlite_BuildGetAuditLogEntriesForWebhookQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		q, _ := buildTestService(t)
		ctx := context.Background()

		exampleWebhook := fakes.BuildFakeWebhook()

		expectedQuery := "SELECT audit_log.id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE json_extract(audit_log.context, '$.webhook_id') = ? ORDER BY audit_log.created_on"
		expectedArgs := []interface{}{
			exampleWebhook.ID,
		}
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForWebhookQuery(ctx, exampleWebhook.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
