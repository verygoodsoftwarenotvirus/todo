package frontend

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_fetchWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleWebhook, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.fetchWebhook(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleWebhook, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.fetchWebhook(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.fetchWebhook(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

/*

func attachWebhookCreationInputToRequest(input *types.WebhookCreationInput) *http.Request {
	form := url.Values{
		webhookCreationInputNameFormKey:    {input.Name},
		webhookCreationInputDetailsFormKey: {input.Details},
	}

	return httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(form.Encode()))
}

func TestService_buildWebhookCreatorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookCreatorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookCreatorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookCreatorView(false)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with base template and error writing to response", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := &testutils.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookCreatorView(true)(res, req)
	})

	T.Run("without base template and error writing to response", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := &testutils.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookCreatorView(false)(res, req)
	})
}

func TestService_parseFormEncodedWebhookCreationInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		expected := fakes.BuildFakeWebhookCreationInput()
		req := attachWebhookCreationInputToRequest(expected)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: expected.BelongsToAccount,
		}

		actual := s.parseFormEncodedWebhookCreationInput(ctx, req, sessionCtxData)
		assert.Equal(t, expected, actual)
	})

	T.Run("with error extracting form from request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleInput := fakes.BuildFakeWebhookCreationInput()

		badBody := &testutils.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/test", badBody)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: exampleInput.BelongsToAccount,
		}

		actual := s.parseFormEncodedWebhookCreationInput(ctx, req, sessionCtxData)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleInput := &types.WebhookCreationInput{}
		req := attachWebhookCreationInputToRequest(exampleInput)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: exampleInput.BelongsToAccount,
		}

		actual := s.parseFormEncodedWebhookCreationInput(ctx, req, sessionCtxData)
		assert.Nil(t, actual)
	})
}

func TestService_handleWebhookCreationRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}
		exampleInput.BelongsToAccount = exampleSessionContextData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachWebhookCreationInputToRequest(exampleInput)

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			exampleInput,
			exampleSessionContextData.Requester.ID,
		).Return(exampleWebhook, nil)
		s.dataStore = mockDB

		s.handleWebhookCreationRequest(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := attachWebhookCreationInputToRequest(exampleInput)

		s.handleWebhookCreationRequest(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}
		exampleInput.BelongsToAccount = exampleSessionContextData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachWebhookCreationInputToRequest(&types.WebhookCreationInput{})

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			exampleInput,
			exampleSessionContextData.Requester.ID,
		).Return(exampleWebhook, nil)
		s.dataStore = mockDB

		s.handleWebhookCreationRequest(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error creating webhook in database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}
		exampleInput.BelongsToAccount = exampleSessionContextData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachWebhookCreationInputToRequest(exampleInput)

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			exampleInput,
			exampleSessionContextData.Requester.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.dataStore = mockDB

		s.handleWebhookCreationRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

*/

func TestService_buildWebhookEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleWebhook, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleWebhook, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookEditorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhookEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_fetchWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		exampleWebhookList := fakes.BuildFakeWebhookList()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.fetchWebhooks(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleWebhookList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.fetchWebhooks(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.fetchWebhooks(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildWebhooksTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhooksTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhooksTableView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhooksTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.buildWebhooksTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
