package frontend

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func attachItemCreationInputToRequest(input *types.ItemCreationInput) *http.Request {
	form := url.Values{
		itemCreationInputNameFormKey:    {input.Name},
		itemCreationInputDetailsFormKey: {input.Details},
	}

	return httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
}

func attachItemUpdateInputToRequest(input *types.ItemUpdateInput) *http.Request {
	form := url.Values{
		itemCreationInputNameFormKey:    {input.Name},
		itemCreationInputDetailsFormKey: {input.Details},
	}

	return httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
}

func TestService_fetchItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleItem := fakes.BuildFakeItem()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleItem, nil)
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.fetchItem(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleItem, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.fetchItem(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleItem := fakes.BuildFakeItem()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.fetchItem(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})
}

func TestService_buildItemCreatorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemCreatorView(true)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemCreatorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemCreatorView(false)(res, req)

		assert.Equal(t, http.StatusSeeOther, res.Code)
	})

	T.Run("with base template and error writing to response", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := &testutil.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemCreatorView(true)(res, req)
	})

	T.Run("without base template and error writing to response", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := &testutil.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemCreatorView(false)(res, req)
	})
}

func TestService_parseFormEncodedItemCreationInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		expected := fakes.BuildFakeItemCreationInput()
		req := attachItemCreationInputToRequest(expected)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: expected.BelongsToAccount,
		}

		actual := s.parseFormEncodedItemCreationInput(ctx, req, sessionCtxData)
		assert.Equal(t, expected, actual)
	})

	T.Run("with error extracting form from request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleInput := fakes.BuildFakeItemCreationInput()

		badBody := &testutil.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodGet, "/test", badBody)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: exampleInput.BelongsToAccount,
		}

		actual := s.parseFormEncodedItemCreationInput(ctx, req, sessionCtxData)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleInput := &types.ItemCreationInput{}
		req := attachItemCreationInputToRequest(exampleInput)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: exampleInput.BelongsToAccount,
		}

		actual := s.parseFormEncodedItemCreationInput(ctx, req, sessionCtxData)
		assert.Nil(t, actual)
	})
}

func TestService_handleItemCreationRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}
		exampleInput.BelongsToAccount = exampleSessionContextData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(exampleInput)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			exampleInput,
			exampleSessionContextData.Requester.ID,
		).Return(exampleItem, nil)
		s.dataStore = mockDB

		s.handleItemCreationRequest(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(exampleInput)

		s.handleItemCreationRequest(res, req)

		assert.Equal(t, http.StatusSeeOther, res.Code)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}
		exampleInput.BelongsToAccount = exampleSessionContextData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(&types.ItemCreationInput{})

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			exampleInput,
			exampleSessionContextData.Requester.ID,
		).Return(exampleItem, nil)
		s.dataStore = mockDB

		s.handleItemCreationRequest(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error creating item in database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}
		exampleInput.BelongsToAccount = exampleSessionContextData.ActiveAccountID

		res := httptest.NewRecorder()
		req := attachItemCreationInputToRequest(exampleInput)

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			exampleInput,
			exampleSessionContextData.Requester.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.dataStore = mockDB

		s.handleItemCreationRequest(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildItemEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleItem, nil)
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleItem, nil)
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemEditorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemEditorView(true)(res, req)

		assert.Equal(t, http.StatusSeeOther, res.Code)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleItem := fakes.BuildFakeItem()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return((*types.Item)(nil), errors.New("blah"))
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})
}

func TestService_fetchItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		exampleItemList := fakes.BuildFakeItemList()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.fetchItems(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleItemList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.fetchItems(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		actual, err := s.fetchItems(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildItemsTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItemList := fakes.BuildFakeItemList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemsTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItemList := fakes.BuildFakeItemList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildItemsTableView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_parseFormEncodedItemUpdateInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleItem := fakes.BuildFakeItem()
		expected := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)
		sessionCtxData := &types.SessionContextData{
			ActiveAccountID: expected.BelongsToAccount,
		}

		req := attachItemUpdateInputToRequest(expected)

		actual := s.parseFormEncodedItemUpdateInput(ctx, req, sessionCtxData)
		assert.Equal(t, expected, actual)
	})
}

func TestService_handleItemUpdateRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
		).Return(exampleItem, nil)

		mockDB.ItemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			exampleItem,
			exampleSessionContextData.Requester.ID,
			[]*types.FieldChangeSummary(nil),
		).Return(nil)
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		res := httptest.NewRecorder()
		req := attachItemUpdateInputToRequest(exampleInput)

		s.handleItemUpdateRequest(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})
}

func TestService_handleItemDeletionRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleItem := fakes.BuildFakeItem()
		exampleItemList := fakes.BuildFakeItemList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		mockDB := database.BuildMockDatabase()
		mockDB.ItemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			exampleItem.ID,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.ID,
		).Return(nil)
		s.dataStore = mockDB

		mockDB.ItemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		s.dataStore = mockDB

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			itemIDURLParamKey,
			"item",
		).Return(func(req *http.Request) uint64 {
			return exampleItem.ID
		})
		s.routeParamManager = rpm

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {

			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/items", nil)

		s.handleItemDeletionRequest(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})
}
