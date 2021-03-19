package postgres

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildGetBatchOfAPIClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account FROM api_clients WHERE api_clients.id > $1 AND api_clients.id < $2"
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

func TestPostgres_BuildGetAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.client_id = $1"
		expectedArgs := []interface{}{
			exampleAPIClient.ClientID,
		}
		actualQuery, actualArgs := q.BuildGetAPIClientByClientIDQuery(exampleAPIClient.ClientID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllAPIClientsCountQuery(T *testing.T) {
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

func TestPostgres_BuildGetAPIClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT api_clients.id, api_clients.external_id, api_clients.name, api_clients.client_id, api_clients.secret_key, api_clients.created_on, api_clients.last_updated_on, api_clients.archived_on, api_clients.belongs_to_account, (SELECT COUNT(api_clients.id) FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = $1) as total_count, (SELECT COUNT(api_clients.id) FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = $2 AND api_clients.created_on > $3 AND api_clients.created_on < $4 AND api_clients.last_updated_on > $5 AND api_clients.last_updated_on < $6) as filtered_count FROM api_clients WHERE api_clients.archived_on IS NULL AND api_clients.belongs_to_account = $7 AND api_clients.created_on > $8 AND api_clients.created_on < $9 AND api_clients.last_updated_on > $10 AND api_clients.last_updated_on < $11 GROUP BY api_clients.id LIMIT 20 OFFSET 180"
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

func TestPostgres_BuildCreateAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleAPIClientInput := fakes.BuildFakeAPIClientCreationInputFromClient(exampleAPIClient)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleAPIClient.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO api_clients (external_id,name,client_id,secret_key,belongs_to_account) VALUES ($1,$2,$3,$4,$5) RETURNING id"
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

func TestPostgres_BuildUpdateAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "UPDATE api_clients SET client_id = $1, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $2 AND id = $3"
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

func TestPostgres_BuildArchiveAPIClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()

		expectedQuery := "UPDATE api_clients SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $1 AND id = $2"
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
