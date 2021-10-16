package workers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockpublishers "gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

func TestProvidePreWritesWorker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}
		dbManager := &database.MockDatabase{}
		postArchivesPublisher := &mockpublishers.Publisher{}
		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		actual, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with error providing search index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}
		dbManager := &database.MockDatabase{}
		postArchivesPublisher := &mockpublishers.Publisher{}
		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, errors.New("blah")
		}

		actual, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})
}

func TestPreWritesWorker_HandleMessage(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}
		dbManager := database.BuildMockDatabase()
		postArchivesPublisher := &mockpublishers.Publisher{}
		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, []byte("} bad JSON lol")))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with ItemDataType", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.ItemDataType,
			Item:     fakes.BuildFakeItemDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		expectedItem := fakes.BuildFakeItem()

		dbManager := database.BuildMockDatabase()
		dbManager.ItemDataManager.On(
			"CreateItem",
			testutils.ContextMatcher,
			body.Item,
		).Return(expectedItem, nil)

		searchIndexManager := &mocksearch.IndexManager{}
		searchIndexManager.On(
			"Index",
			testutils.ContextMatcher,
			expectedItem.ID,
			expectedItem,
		).Return(nil)

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return searchIndexManager, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}
		postArchivesPublisher.On(
			"Publish",
			testutils.ContextMatcher,
			mock.MatchedBy(func(message *types.DataChangeMessage) bool { return true }),
		).Return(nil)

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.NoError(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher, searchIndexManager)
	})

	T.Run("with ItemDataType and error writing", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.ItemDataType,
			Item:     fakes.BuildFakeItemDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		dbManager := database.BuildMockDatabase()
		dbManager.ItemDataManager.On(
			"CreateItem",
			testutils.ContextMatcher,
			body.Item,
		).Return((*types.Item)(nil), errors.New("blah"))

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with ItemDataType and error updating search index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.ItemDataType,
			Item:     fakes.BuildFakeItemDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		expectedItem := fakes.BuildFakeItem()

		dbManager := database.BuildMockDatabase()
		dbManager.ItemDataManager.On(
			"CreateItem",
			testutils.ContextMatcher,
			body.Item,
		).Return(expectedItem, nil)

		searchIndexManager := &mocksearch.IndexManager{}
		searchIndexManager.On(
			"Index",
			testutils.ContextMatcher,
			expectedItem.ID,
			expectedItem,
		).Return(errors.New("blah"))

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return searchIndexManager, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher, searchIndexManager)
	})

	T.Run("with ItemDataType and error publishing data change message", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.ItemDataType,
			Item:     fakes.BuildFakeItemDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		expectedItem := fakes.BuildFakeItem()

		dbManager := database.BuildMockDatabase()
		dbManager.ItemDataManager.On(
			"CreateItem",
			testutils.ContextMatcher,
			body.Item,
		).Return(expectedItem, nil)

		searchIndexManager := &mocksearch.IndexManager{}
		searchIndexManager.On(
			"Index",
			testutils.ContextMatcher,
			expectedItem.ID,
			expectedItem,
		).Return(nil)

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return searchIndexManager, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}
		postArchivesPublisher.On(
			"Publish",
			testutils.ContextMatcher,
			mock.MatchedBy(func(message *types.DataChangeMessage) bool { return true }),
		).Return(errors.New("blah"))

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher, searchIndexManager)
	})

	T.Run("with WebhookDataType", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.WebhookDataType,
			Webhook:  fakes.BuildFakeWebhookDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		expectedWebhook := fakes.BuildFakeWebhook()

		dbManager := database.BuildMockDatabase()
		dbManager.WebhookDataManager.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			body.Webhook,
		).Return(expectedWebhook, nil)

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}
		postArchivesPublisher.On(
			"Publish",
			testutils.ContextMatcher,
			mock.MatchedBy(func(message *types.DataChangeMessage) bool { return true }),
		).Return(nil)

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.NoError(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with WebhookDataType and error writing", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.WebhookDataType,
			Webhook:  fakes.BuildFakeWebhookDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		dbManager := database.BuildMockDatabase()
		dbManager.WebhookDataManager.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			body.Webhook,
		).Return((*types.Webhook)(nil), errors.New("blah"))

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with WebhookDataType and error publishing data change message", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType: types.WebhookDataType,
			Webhook:  fakes.BuildFakeWebhookDatabaseCreationInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		expectedWebhook := fakes.BuildFakeWebhook()

		dbManager := database.BuildMockDatabase()
		dbManager.WebhookDataManager.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			body.Webhook,
		).Return(expectedWebhook, nil)

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}
		postArchivesPublisher.On(
			"Publish",
			testutils.ContextMatcher,
			mock.MatchedBy(func(message *types.DataChangeMessage) bool { return true }),
		).Return(errors.New("blah"))

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with UserMembershipDataType", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType:       types.UserMembershipDataType,
			UserMembership: fakes.BuildFakeAddUserToAccountInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		dbManager := database.BuildMockDatabase()
		dbManager.AccountUserMembershipDataManager.On(
			"AddUserToAccount",
			testutils.ContextMatcher,
			body.UserMembership,
		).Return(nil)

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}
		postArchivesPublisher.On(
			"Publish",
			testutils.ContextMatcher,
			mock.MatchedBy(func(message *types.DataChangeMessage) bool { return true }),
		).Return(nil)

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.NoError(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with UserMembershipDataType and error writing", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType:       types.UserMembershipDataType,
			UserMembership: fakes.BuildFakeAddUserToAccountInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		dbManager := database.BuildMockDatabase()
		dbManager.AccountUserMembershipDataManager.On(
			"AddUserToAccount",
			testutils.ContextMatcher,
			mock.MatchedBy(func(input *types.AddUserToAccountInput) bool { return true }),
		).Return(errors.New("blah"))

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})

	T.Run("with UserMembershipDataType and error publishing data change message", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		client := &http.Client{}

		body := &types.PreWriteMessage{
			DataType:       types.UserMembershipDataType,
			UserMembership: fakes.BuildFakeAddUserToAccountInput(),
		}
		examplePayload, err := json.Marshal(body)
		require.NoError(t, err)

		dbManager := database.BuildMockDatabase()
		dbManager.AccountUserMembershipDataManager.On(
			"AddUserToAccount",
			testutils.ContextMatcher,
			mock.MatchedBy(func(input *types.AddUserToAccountInput) bool { return true }),
		).Return(nil)

		searchIndexLocation := search.IndexPath(t.Name())
		searchIndexProvider := func(context.Context, logging.Logger, *http.Client, search.IndexPath, search.IndexName, ...string) (search.IndexManager, error) {
			return nil, nil
		}

		postArchivesPublisher := &mockpublishers.Publisher{}
		postArchivesPublisher.On(
			"Publish",
			testutils.ContextMatcher,
			mock.MatchedBy(func(message *types.DataChangeMessage) bool { return true }),
		).Return(errors.New("blah"))

		worker, err := ProvidePreWritesWorker(
			ctx,
			logger,
			client,
			dbManager,
			postArchivesPublisher,
			searchIndexLocation,
			searchIndexProvider,
		)
		require.NotNil(t, worker)
		require.NoError(t, err)

		assert.Error(t, worker.HandleMessage(ctx, examplePayload))

		mock.AssertExpectationsForObjects(t, dbManager, postArchivesPublisher)
	})
}
