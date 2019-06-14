package postgres

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_ExportData(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, mockDB := buildTestService(t)

		ctx := context.Background()
		exampleUser := &models.User{ID: 123}
		expected := &models.DataExport{
			User: *exampleUser,
			Items: []models.Item{
				{ID: 123},
			},
			OAuth2Clients: []models.OAuth2Client{
				{ID: 123},
			},
			Webhooks: []models.Webhook{
				{ID: 123},
			},
		}

		// prep for items query
		expectedItemListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = $1"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedItemListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowFromItem(&expected.Items[0]),
			)

		// prep for OAuth2 clients query
		expectedOAuth2ClientListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedOAuth2ClientListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowFromOAuth2Client(&expected.OAuth2Clients[0]),
			)

		// prep for webhooks query
		expectedWebhookListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedWebhookListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowFromWebhook(&expected.Webhooks[0]),
			)

		actual, err := p.ExportData(ctx, exampleUser)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("with error fetching items", func(t *testing.T) {
		p, mockDB := buildTestService(t)

		ctx := context.Background()
		exampleUser := &models.User{ID: 123}

		// prep for items query
		expectedItemListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = $1"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedItemListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnError(errors.New("blah"))

		actual, err := p.ExportData(ctx, exampleUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error fetching oauth2 clients", func(t *testing.T) {
		p, mockDB := buildTestService(t)

		ctx := context.Background()
		exampleUser := &models.User{ID: 123}
		expected := &models.DataExport{
			User: *exampleUser,
			Items: []models.Item{
				{ID: 123},
			},
			OAuth2Clients: []models.OAuth2Client{
				{ID: 123},
			},
			Webhooks: []models.Webhook{
				{ID: 123},
			},
		}

		// prep for items query
		expectedItemListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = $1"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedItemListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowFromItem(&expected.Items[0]),
			)

		// prep for OAuth2 clients query
		expectedOAuth2ClientListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedOAuth2ClientListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnError(errors.New("blah"))

		actual, err := p.ExportData(ctx, exampleUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error fetching webhooks", func(t *testing.T) {
		p, mockDB := buildTestService(t)

		ctx := context.Background()
		exampleUser := &models.User{ID: 123}
		expected := &models.DataExport{
			User: *exampleUser,
			Items: []models.Item{
				{ID: 123},
			},
			OAuth2Clients: []models.OAuth2Client{
				{ID: 123},
			},
			Webhooks: []models.Webhook{
				{ID: 123},
			},
		}

		// prep for items query
		expectedItemListQuery := "SELECT id, name, details, created_on, updated_on, archived_on, belongs_to FROM items WHERE archived_on IS NULL AND belongs_to = $1"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedItemListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowFromItem(&expected.Items[0]),
			)

		// prep for OAuth2 clients query
		expectedOAuth2ClientListQuery := "SELECT id, name, client_id, scopes, redirect_uri, client_secret, created_on, updated_on, archived_on, belongs_to FROM oauth2_clients WHERE archived_on IS NULL"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedOAuth2ClientListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnRows(
				buildMockRowFromOAuth2Client(&expected.OAuth2Clients[0]),
			)

		// prep for webhooks query
		expectedWebhookListQuery := "SELECT id, name, content_type, url, method, events, data_types, topics, created_on, updated_on, archived_on, belongs_to FROM webhooks WHERE archived_on IS NULL"
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedWebhookListQuery)).
			WithArgs(exampleUser.ID).
			WillReturnError(errors.New("blah"))

		actual, err := p.ExportData(ctx, exampleUser)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

}
