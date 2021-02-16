package postgres

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_BuildGetBatchOfDelegatedClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		beginID, endID := uint64(1), uint64(1000)

		expectedQuery := "SELECT delegated_clients.id, delegated_clients.external_id, delegated_clients.name, delegated_clients.client_id, delegated_clients.secret_key, delegated_clients.created_on, delegated_clients.last_updated_on, delegated_clients.archived_on, delegated_clients.belongs_to_user FROM delegated_clients WHERE delegated_clients.id > $1 AND delegated_clients.id < $2"
		expectedArgs := []interface{}{
			beginID,
			endID,
		}
		actualQuery, actualArgs := q.BuildGetBatchOfDelegatedClientsQuery(beginID, endID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetDelegatedClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		expectedQuery := "SELECT delegated_clients.id, delegated_clients.external_id, delegated_clients.name, delegated_clients.client_id, delegated_clients.secret_key, delegated_clients.created_on, delegated_clients.last_updated_on, delegated_clients.archived_on, delegated_clients.belongs_to_user FROM delegated_clients WHERE delegated_clients.archived_on IS NULL AND delegated_clients.client_id = $1"
		expectedArgs := []interface{}{
			exampleDelegatedClient.ClientID,
		}
		actualQuery, actualArgs := q.BuildGetDelegatedClientQuery(exampleDelegatedClient.ClientID)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildGetAllDelegatedClientsCountQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		expectedQuery := "SELECT COUNT(delegated_clients.id) FROM delegated_clients WHERE delegated_clients.archived_on IS NULL"
		actualQuery := q.BuildGetAllDelegatedClientsCountQuery()

		assertArgCountMatchesQuery(t, actualQuery, []interface{}{})
		assert.Equal(t, expectedQuery, actualQuery)
	})
}

func TestPostgres_BuildGetDelegatedClientsQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		filter := fakes.BuildFleshedOutQueryFilter()

		expectedQuery := "SELECT delegated_clients.id, delegated_clients.external_id, delegated_clients.name, delegated_clients.client_id, delegated_clients.secret_key, delegated_clients.created_on, delegated_clients.last_updated_on, delegated_clients.archived_on, delegated_clients.belongs_to_user, (SELECT COUNT(delegated_clients.id) FROM delegated_clients WHERE delegated_clients.archived_on IS NULL AND delegated_clients.belongs_to_user = $1) as total_count, (SELECT COUNT(delegated_clients.id) FROM delegated_clients WHERE delegated_clients.archived_on IS NULL AND delegated_clients.belongs_to_user = $2 AND delegated_clients.created_on > $3 AND delegated_clients.created_on < $4 AND delegated_clients.last_updated_on > $5 AND delegated_clients.last_updated_on < $6) as filtered_count FROM delegated_clients WHERE delegated_clients.archived_on IS NULL AND delegated_clients.belongs_to_user = $7 AND delegated_clients.created_on > $8 AND delegated_clients.created_on < $9 AND delegated_clients.last_updated_on > $10 AND delegated_clients.last_updated_on < $11 GROUP BY delegated_clients.id LIMIT 20 OFFSET 180"
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
		actualQuery, actualArgs := q.BuildGetDelegatedClientsQuery(exampleUser.ID, filter)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildCreateDelegatedClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()
		exampleDelegatedClientInput := fakes.BuildFakeDelegatedClientCreationInputFromClient(exampleDelegatedClient)

		exIDGen := &querybuilding.MockExternalIDGenerator{}
		exIDGen.On("NewExternalID").Return(exampleDelegatedClient.ExternalID)
		q.externalIDGenerator = exIDGen

		expectedQuery := "INSERT INTO delegated_clients (external_id,name,client_id,secret_key,belongs_to_user) VALUES ($1,$2,$3,$4,$5)"
		expectedArgs := []interface{}{
			exampleDelegatedClient.ExternalID,
			exampleDelegatedClient.Name,
			exampleDelegatedClient.ClientID,
			exampleDelegatedClient.ClientSecret,
			exampleDelegatedClient.BelongsToUser,
		}
		actualQuery, actualArgs := q.BuildCreateDelegatedClientQuery(exampleDelegatedClientInput)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)

		mock.AssertExpectationsForObjects(t, exIDGen)
	})
}

func TestPostgres_BuildUpdateDelegatedClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		expectedQuery := "UPDATE delegated_clients SET client_id = $1, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $2 AND id = $3"
		expectedArgs := []interface{}{
			exampleDelegatedClient.ClientID,
			exampleDelegatedClient.BelongsToUser,
			exampleDelegatedClient.ID,
		}
		actualQuery, actualArgs := q.BuildUpdateDelegatedClientQuery(exampleDelegatedClient)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}

func TestPostgres_BuildArchiveDelegatedClientQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		expectedQuery := "UPDATE delegated_clients SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2"
		expectedArgs := []interface{}{
			exampleDelegatedClient.BelongsToUser,
			exampleDelegatedClient.ID,
		}
		actualQuery, actualArgs := q.BuildArchiveDelegatedClientQuery(exampleDelegatedClient.ID, exampleDelegatedClient.BelongsToUser)

		assertArgCountMatchesQuery(t, actualQuery, actualArgs)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, expectedArgs, actualArgs)
	})
}
