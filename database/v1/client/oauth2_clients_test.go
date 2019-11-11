package dbclient

import (
	"context"
	"errors"
	"fmt"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)
		expected := &models.OAuth2Client{}

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)
		expected := (*models.OAuth2Client)(nil)

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := "CLIENT_ID"
		c, mockDB := buildTestClient()
		expected := &models.OAuth2Client{}

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleClientID).Return(expected, nil)

		actual, err := c.GetOAuth2ClientByClientID(context.Background(), exampleClientID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		exampleClientID := "CLIENT_ID"
		c, mockDB := buildTestClient()
		expected := (*models.OAuth2Client)(nil)

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleClientID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2ClientByClientID(context.Background(), exampleClientID)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUserID := uint64(123)
		expected := uint64(123)
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientCount", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2ClientCount(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		exampleUserID := uint64(123)
		expected := uint64(123)
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientCount", mock.Anything, mock.AnythingOfType("*models.QueryFilter"), exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2ClientCount(context.Background(), nil, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		exampleUserID := uint64(123)
		expected := uint64(0)
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientCount", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2ClientCount(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		expected := uint64(123)
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllOAuth2ClientCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		var expected []*models.OAuth2Client
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2Clients", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllOAuth2Clients(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		exampleUserID := uint64(123)
		expected := &models.OAuth2ClientList{}
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		c, mockDB := buildTestClient()
		exampleUserID := uint64(123)
		expected := &models.OAuth2ClientList{}
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, (*models.QueryFilter)(nil), exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2Clients(context.Background(), nil, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		c, mockDB := buildTestClient()
		exampleUserID := uint64(123)
		expected := (*models.OAuth2ClientList)(nil)
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, models.DefaultQueryFilter(), exampleUserID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2Clients(context.Background(), models.DefaultQueryFilter(), exampleUserID)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		c, mockDB := buildTestClient()
		expected := &models.OAuth2Client{}
		exampleInput := &models.OAuth2ClientCreationInput{}
		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateOAuth2Client(context.Background(), exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		c, mockDB := buildTestClient()
		expected := (*models.OAuth2Client)(nil)
		exampleInput := &models.OAuth2ClientCreationInput{}
		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(expected, errors.New("blah"))

		actual, err := c.CreateOAuth2Client(context.Background(), exampleInput)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		example := &models.OAuth2Client{}
		var expected error
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("UpdateOAuth2Client", mock.Anything, example).Return(expected)

		actual := c.UpdateOAuth2Client(context.Background(), example)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)
		var expected error
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected)

		actual := c.ArchiveOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		exampleClientID := uint64(321)
		exampleUserID := uint64(123)
		expected := fmt.Errorf("blah")
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected)

		actual := c.ArchiveOAuth2Client(context.Background(), exampleClientID, exampleUserID)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}
