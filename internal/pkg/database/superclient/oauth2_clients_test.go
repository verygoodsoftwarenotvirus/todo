package superclient

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		c, _, mockDB := buildTestClient(t)
		mockDB.OAuth2ClientDataManager.On("GetOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(exampleOAuth2Client, nil)

		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expected := (*types.OAuth2Client)(nil)

		c, _, mockDB := buildTestClient(t)
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

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		c, _, mockDB := buildTestClient(t)
		mockDB.OAuth2ClientDataManager.On("GetOAuth2ClientByClientID", mock.Anything, exampleOAuth2Client.ClientID).Return(exampleOAuth2Client, nil)

		actual, err := c.GetOAuth2ClientByClientID(ctx, exampleOAuth2Client.ClientID)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		c, _, mockDB := buildTestClient(t)
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

		c, _, mockDB := buildTestClient(t)
		mockDB.OAuth2ClientDataManager.On("GetTotalOAuth2ClientCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetTotalOAuth2ClientCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetOAuth2ClientsForUser(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _, mockDB := buildTestClient(t)
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		filter := types.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, exampleUser.ID, filter).Return(exampleOAuth2ClientList, nil)

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _, mockDB := buildTestClient(t)
		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()
		filter := (*types.QueryFilter)(nil)

		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, exampleUser.ID, filter).Return(exampleOAuth2ClientList, nil)

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2ClientList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _, mockDB := buildTestClient(t)
		exampleOAuth2ClientList := (*types.OAuth2ClientList)(nil)
		filter := types.DefaultQueryFilter()

		mockDB.OAuth2ClientDataManager.On("GetOAuth2Clients", mock.Anything, exampleUser.ID, filter).Return(exampleOAuth2ClientList, errors.New("blah"))

		actual, err := c.GetOAuth2Clients(ctx, exampleUser.ID, filter)
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

		c, _, mockDB := buildTestClient(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		mockDB.OAuth2ClientDataManager.On("CreateOAuth2Client", mock.Anything, exampleInput).Return(exampleOAuth2Client, nil)

		actual, err := c.CreateOAuth2Client(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _, mockDB := buildTestClient(t)

		expected := (*types.OAuth2Client)(nil)
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

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

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		var expected error
		c, _, mockDB := buildTestClient(t)
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

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		var expected error
		c, _, mockDB := buildTestClient(t)
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(expected)

		actual := c.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.NoError(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error returned from querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		expected := fmt.Errorf("blah")
		c, _, mockDB := buildTestClient(t)
		mockDB.OAuth2ClientDataManager.On("ArchiveOAuth2Client", mock.Anything, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser).Return(expected)

		actual := c.ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID, exampleOAuth2Client.BelongsToUser)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
