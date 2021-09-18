package frontend

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_fetchItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleItem, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.service.fetchItem(s.ctx, req, s.sessionCtxData)
		assert.Equal(t, exampleItem, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.service.fetchItem(s.ctx, req, s.sessionCtxData)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.service.fetchItem(s.ctx, req, s.sessionCtxData)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func attachItemCreationInputToRequest(input *types.ItemDatabaseCreationInput) *http.Request {
	form := url.Values{
		itemCreationInputNameFormKey:    {anyToString(input.Name)},
		itemCreationInputDetailsFormKey: {anyToString(input.Details)},
	}

	return httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
}

func TestService_buildItemCreatorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemCreatorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemCreatorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemCreatorView(false)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with base template and error writing to response", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := &testutils.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemCreatorView(true)(res, req)
	})

	T.Run("without base template and error writing to response", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := &testutils.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemCreatorView(false)(res, req)
	})
}

func TestService_parseFormEncodedItemCreationInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		expected := fakes.BuildFakeItemDatabaseCreationInput()
		expected.ID = ""
		expected.BelongsToAccount = s.exampleAccount.ID
		req := attachItemCreationInputToRequest(expected)

		actual := s.service.parseFormEncodedItemCreationInput(s.ctx, req, s.sessionCtxData)
		assert.Equal(t, expected, actual)
	})

	T.Run("with error extracting form from request", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		badBody := &testutils.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/test", badBody)

		actual := s.service.parseFormEncodedItemCreationInput(s.ctx, req, s.sessionCtxData)
		assert.Nil(t, actual)
	})
}

func TestService_handleItemCreationRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)
		exampleInput.ID = ""
		exampleInput.BelongsToAccount = s.sessionCtxData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(exampleInput)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"CreateItem",
			testutils.ContextMatcher,
			exampleInput,
		).Return(exampleItem, nil)
		s.service.dataStore = mockDB

		s.service.handleItemCreationRequest(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)
		assert.NotEmpty(t, res.Header().Get(htmxRedirectionHeader))

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)
		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(exampleInput)

		s.service.handleItemCreationRequest(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error creating item in database", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemDatabaseCreationInputFromItem(exampleItem)
		exampleInput.ID = ""
		exampleInput.BelongsToAccount = s.sessionCtxData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(exampleInput)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"CreateItem",
			testutils.ContextMatcher,
			exampleInput,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		s.service.handleItemCreationRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildItemEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleItem, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleItem, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemEditorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_fetchItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItemList := fakes.BuildFakeItemList()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.service.fetchItems(s.ctx, req, s.sessionCtxData)
		assert.Equal(t, exampleItemList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.service.fetchItems(s.ctx, req, s.sessionCtxData)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.service.fetchItems(s.ctx, req, s.sessionCtxData)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildItemsTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItemList := fakes.BuildFakeItemList()
		for _, item := range exampleItemList.Items {
			item.BelongsToAccount = s.exampleAccount.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemsTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItemList := fakes.BuildFakeItemList()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemsTableView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemsTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.service.buildItemsTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func attachItemUpdateInputToRequest(input *types.ItemUpdateInput) *http.Request {
	form := url.Values{
		itemUpdateInputNameFormKey:    {anyToString(input.Name)},
		itemUpdateInputDetailsFormKey: {anyToString(input.Details)},
	}

	return httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
}

func TestService_parseFormEncodedItemUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		expected := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		req := attachItemUpdateInputToRequest(expected)

		actual := s.service.parseFormEncodedItemUpdateInput(s.ctx, req, s.sessionCtxData)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid form", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		badBody := &testutils.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/test", badBody)

		actual := s.service.parseFormEncodedItemUpdateInput(s.ctx, req, s.sessionCtxData)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input attached to valid form", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleInput := &types.ItemUpdateInput{}

		req := attachItemUpdateInputToRequest(exampleInput)

		actual := s.service.parseFormEncodedItemUpdateInput(s.ctx, req, s.sessionCtxData)
		assert.Nil(t, actual)
	})
}

func TestService_handleItemUpdateRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleItem, nil)

		mockDB.ItemDataManager.On(
			"UpdateItem",
			testutils.ContextMatcher,
			exampleItem,
		).Return(nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := attachItemUpdateInputToRequest(exampleInput)

		s.service.handleItemUpdateRequest(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := attachItemUpdateInputToRequest(exampleInput)

		s.service.handleItemUpdateRequest(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleInput := &types.ItemUpdateInput{}

		res := httptest.NewRecorder()
		req := attachItemUpdateInputToRequest(exampleInput)

		s.service.handleItemUpdateRequest(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := attachItemUpdateInputToRequest(exampleInput)

		s.service.handleItemUpdateRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error updating data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleItem, nil)

		mockDB.ItemDataManager.On(
			"UpdateItem",
			testutils.ContextMatcher,
			exampleItem,
		).Return(errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := attachItemUpdateInputToRequest(exampleInput)

		s.service.handleItemUpdateRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_handleItemArchiveRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		exampleItemList := fakes.BuildFakeItemList()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"ArchiveItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(nil)
		s.service.dataStore = mockDB

		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/items", nil)

		s.service.handleItemArchiveRequest(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/items", nil)

		s.service.handleItemArchiveRequest(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error archiving item", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"ArchiveItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/items", nil)

		s.service.handleItemArchiveRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error retrieving new list of items", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = s.exampleAccount.ID
		s.service.itemIDFetcher = func(*http.Request) string {
			return exampleItem.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"ArchiveItem",
			testutils.ContextMatcher,
			exampleItem.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(nil)
		s.service.dataStore = mockDB

		mockDB.ItemDataManager.On(
			"GetItems",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/items", nil)

		s.service.handleItemArchiveRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
