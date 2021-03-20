package mariadb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestMariaDB_BuildGetBatchOfAPIClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.permissions, api_clients.admin_permissions, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account FROM api_clients WHERE api_clients.id > ? AND api_clients.id < ?"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfAPIClientsQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.permissions, api_clients.admin_permissions, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.client_id = ?"
		expectedArgs := []interface{}{
			exampleAPIClient.ClientID,
		}
		actualQuery, actualArgs := q.BuildGetAPIClientByClientIDQuery(exampleAPIClient.ClientID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAPIClientByDatabaseIDQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.permissions, api_clients.admin_permissions, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = ? AND api_clients.id = ?"
		expectedArgs := []interface{}{
			exampleAPIClient.BelongsToAccount,
			exampleAPIClient.ID,
		}
		actualQuery, actualArgs := q.BuildGetAPIClientByDatabaseIDQuery(exampleAPIClient.ID, exampleAPIClient.BelongsToAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAllAPIClientsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(api_clients.id) FROM api_clients WHERE api_clients.archived_on IS NULL"
		actualQuery := q.BuildGetAllAPIClientsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestMariaDB_BuildGetAPIClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.permissions, api_clients.admin_permissions, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account, (SELECT COUNT(api_clients.id) FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = ?) as total_count, (SELECT COUNT(api_clients.id) FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = ? AND api_clients.created_on > ? AND api_clients.created_on < ? AND api_clients.last_updated_on > ? AND api_clients.last_updated_on < ?) as filtered_count FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = ? AND api_clients.created_on > ? AND api_clients.created_on < ? AND api_clients.last_updated_on > ? AND api_clients.last_updated_on < ? GROUP BY api_clients.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetAPIClientsQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildCreateAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleAPIClient.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO api_clients (external_id,name,client_id,secret_key,belongs_to_account) VALUES (?,?,?,?,?)"
		expectedArgs := []interface{}{
			exampleAPIClient.ExternalID,
			exampleAPIClient.Name,
			exampleAPIClient.ClientID,
			exampleAPIClient.ClientSecret,
			exampleAPIClient.BelongsToAccount,
		}
		actualQuery, actualArgs := q.BuildCreateAPIClientQuery(exampleAPIClientInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestMariaDB_BuildUpdateAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "UPDATE api_clients SET client_id = ?, last_updated_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleAPIClient.ClientID,
			exampleAPIClient.BelongsToAccount,
			exampleAPIClient.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateAPIClientQuery(exampleAPIClient)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildArchiveAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "UPDATE api_clients SET last_updated_on = UNIX_TIMESTAMP(), archived_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND belongs_to_account = ? AND id = ?"
		expectedArgs := []interface{}{
			exampleAPIClient.BelongsToAccount,
			exampleAPIClient.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveAPIClientQuery(exampleAPIClient.ID, exampleAPIClient.BelongsToAccount)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestMariaDB_BuildGetAuditLogEntriesForAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := fmt.Sprintf("SELECT audit_log.id, audit_log.external_id, audit_log.event_type, audit_log.context, audit_log.created_on FROM audit_log WHERE JSON_CONTAINS(audit_log.context, '%d', '$.api_client_id') ORDER BY audit_log.created_on", exampleAPIClient.ID)
		expectedArgs := []interface{}(nil)
		actualQuery, actualArgs := q.BuildGetAuditLogEntriesForAPIClientQuery(exampleAPIClient.ID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
