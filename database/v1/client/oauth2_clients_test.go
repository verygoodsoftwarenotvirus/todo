package dbclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"testing"
)

func TestClient_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)
		expected := &models.OAuth2Client{}

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.
			On("GetOAuth2Client", mock.Anything, exampleClientID, exampleUserID).
			Return(expected, nil)

		actual, err := c.GetOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
func TestClient_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := "CLIENT_ID"
		c, mockDB := buildTestClient()
		expected := &models.OAuth2Client{}

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleClientID).
			Return(expected, nil)

		actual, err := c.GetOAuth2ClientByClientID(context.Background(), exampleClientID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
func TestClient_GetOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(123)
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.
			On("GetOAuth2ClientCount", mock.Anything, models.DefaultQueryFilter, exampleUserID).
			Return(uint64(123), nil)

		c.GetOAuth2ClientCount(context.Background(), models.DefaultQueryFilter, exampleUserID)
	})
}
func TestClient_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).
			Return(uint64(123), nil)

		c.GetAllOAuth2ClientCount(context.Background())
	})
}
func TestClient_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2Clients", mock.Anything).
			Return([]*models.OAuth2Client{}, nil)

		c.GetAllOAuth2Clients(context.Background())
	})
}
func TestClient_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		exampleUserID := uint64(123)
		mockDB.OAuth2ClientDataManager.
			On("GetOAuth2Clients", mock.Anything, models.DefaultQueryFilter, exampleUserID).
			Return(&models.OAuth2ClientList{}, nil)

		c.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter, exampleUserID)
	})
}
func TestClient_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		exampleInput := &models.OAuth2ClientCreationInput{}
		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).
			Return(&models.OAuth2Client{}, nil)

		c.CreateOAuth2Client(context.Background(), exampleInput)
	})
}
func TestClient_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		example := &models.OAuth2Client{}
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("UpdateOAuth2Client", mock.Anything, example).
			Return(nil)

		c.UpdateOAuth2Client(context.Background(), example)
	})
}
func TestClient_DeleteOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("DeleteOAuth2Client", mock.Anything, exampleClientID, exampleUserID).
			Return(nil)

		c.DeleteOAuth2Client(context.Background(), exampleClientID, exampleUserID)
	})
}
