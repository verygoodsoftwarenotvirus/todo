package dbclient

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(exampleOAuth2Client, nil)

		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()
		expected := (*models.OAuth2Client)(nil)

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(expected, errors.New("blah"))

		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleOAuth2Client.ClientID).Return(exampleOAuth2Client, nil)

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleOAuth2Client.ClientID).Return(exampleOAuth2Client, errors.New("blah"))

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllOAuth2ClientCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetAllOAuth2ClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, mockDB := buildTestClient()
		exampleOAuth2ClientList := fakemodels.BuildFakeOAuth2ClientList()
		filter := models.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientsForUser", mock.Anything, exampleUser.ID, filter).Return(exampleOAuth2ClientList, nil)

		actual, err := c.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, mockDB := buildTestClient()
		exampleOAuth2ClientList := fakemodels.BuildFakeOAuth2ClientList()
		filter := (*models.QueryFilter)(nil)

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientsForUser", mock.Anything, exampleUser.ID, filter).Return(exampleOAuth2ClientList, nil)

		actual, err := c.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, mockDB := buildTestClient()
		exampleOAuth2ClientList := (*models.OAuth2ClientList)(nil)
		filter := models.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientsForUser", mock.Anything, exampleUser.ID, filter).Return(exampleOAuth2ClientList, errors.New("blah"))

		actual, err := c.GetOAuth2ClientsForUser(ctx, exampleUser.ID, filter)
		assert.Error(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, mockDB := buildTestClient()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()
		exampleInput := fakemodels.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(exampleOAuth2Client, nil)

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, mockDB := buildTestClient()

		expected := (*models.OAuth2Client)(nil)
		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()
		exampleInput := fakemodels.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(expected, errors.New("blah"))

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.Error(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		var expected error
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("UpdateOAuth2Client", mock.Anything, exampleOAuth2Client).Return(expected)

		actual := c.UpdateOAuth2Client(ctx, exampleOAuth2Client)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		var expected error
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(expected)

		actual := c.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakemodels.BuildFakeOAuth2Client()

		expected := fmt.Errorf("blah")
		c, mockDB := buildTestClient()
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(expected)

		actual := c.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
