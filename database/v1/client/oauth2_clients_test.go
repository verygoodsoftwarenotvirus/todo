package dbclient

import (
	"context"
	"errors"
	"fmt"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleClientID := fake.Uint64()
		exampleUserID := fake.Uint64()
		expected := &models.OAuth2Client{}

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2Client(ctx, exampleClientID, exampleUserID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		ctx := context.Background()
		exampleClientID := fake.Uint64()
		exampleUserID := fake.Uint64()
		expected := (*models.OAuth2Client)(nil)

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2Client(ctx, exampleClientID, exampleUserID)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleClientID := "CLIENT_ID"
		c, mockDB := buildTestClient()
		expected := &models.OAuth2Client{}

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleClientID).Return(expected, nil)

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleClientID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		ctx := context.Background()
		exampleClientID := "CLIENT_ID"
		c, mockDB := buildTestClient()
		expected := (*models.OAuth2Client)(nil)

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleClientID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleClientID)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := fake.Uint64()
		c, mockDB := buildTestClient()
		filter := models.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientCount", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2ClientCount(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := fake.Uint64()
		c, mockDB := buildTestClient()
		filter := (*models.QueryFilter)(nil)

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientCount", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2ClientCount(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		ctx := context.Background()
		exampleUserID := fake.Uint64()
		expected := uint64(0)
		c, mockDB := buildTestClient()
		filter := models.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientCount", mock.Anything, filter, exampleUserID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2ClientCount(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		expected := fake.Uint64()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllOAuth2ClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetAllOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		var expected []*models.OAuth2Client
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2Clients", mock.Anything).Return(expected, nil)

		actual, err := c.GetAllOAuth2Clients(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		exampleUserID := fake.Uint64()
		expected := &models.OAuth2ClientList{}
		filter := models.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2Clients(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		exampleUserID := fake.Uint64()
		expected := &models.OAuth2ClientList{}
		filter := (*models.QueryFilter)(nil)

		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, filter, exampleUserID).Return(expected, nil)

		actual, err := c.GetOAuth2Clients(ctx, exampleUserID, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		exampleUserID := fake.Uint64()
		expected := (*models.OAuth2ClientList)(nil)
		filter := models.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, filter, exampleUserID).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2Clients(ctx, exampleUserID, filter)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		expected := &models.OAuth2Client{}
		exampleInput := &models.OAuth2ClientCreationInput{}
		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		ctx := context.Background()
		c, mockDB := buildTestClient()
		expected := (*models.OAuth2Client)(nil)
		exampleInput := &models.OAuth2ClientCreationInput{}
		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(expected, errors.New("blah"))

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		example := &models.OAuth2Client{}
		var expected error
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("UpdateOAuth2Client", mock.Anything, example).Return(expected)

		actual := c.UpdateOAuth2Client(ctx, example)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleClientID := fake.Uint64()
		exampleUserID := fake.Uint64()
		var expected error
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected)

		actual := c.ArchiveOAuth2Client(ctx, exampleClientID, exampleUserID)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		ctx := context.Background()
		exampleClientID := fake.Uint64()
		exampleUserID := fake.Uint64()
		expected := fmt.Errorf("blah")
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleClientID, exampleUserID).Return(expected)

		actual := c.ArchiveOAuth2Client(ctx, exampleClientID, exampleUserID)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}
