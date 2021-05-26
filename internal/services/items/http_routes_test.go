package items

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParseBool(t *testing.T) {
	t.Parallel()

	expectations := map[string]bool{
		"1":      true,
		t.Name(): false,
		"true":   true,
		"troo":   false,
		"t":      true,
		"false":  false,
	}

	for input, expected := range expectations {
		assert.Equal(t, expected, parseBool(input))
	}
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleItem,
		).Return(nil)
		helper.service.search = indexManager

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, unitCounter, indexManager)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input attached", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := &types.ItemCreationInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error creating item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
			helper.exampleUser.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error indexing item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleItem,
		).Return(errors.New("blah"))
		helper.service.search = indexManager

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, unitCounter, indexManager)
	})
}

func TestItemsService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType(&types.Item{}))
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such item in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ItemExists",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(true, nil)
		helper.service.itemDataManager = itemDataManager

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no result in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ItemExists",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(false, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error checking database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ItemExists",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(false, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleItemList := fakes.BuildFakeItemList()

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType(&types.ItemList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType(&types.ItemList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving items from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_SearchHandler(T *testing.T) {
	T.Parallel()

	exampleQuery := "whatever"
	exampleLimit := uint8(123)
	exampleItemList := fakes.BuildFakeItemList()
	exampleItemIDs := []uint64{}
	for _, x := range exampleItemList.Items {
		exampleItemIDs = append(exampleItemIDs, x.ID)
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return(exampleItemList.Items, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType([]*types.Item{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error conducting search", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{types.SearchQueryKey: []string{exampleQuery}}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return([]uint64{}, errors.New("blah"))
		helper.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, encoderDecoder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return([]*types.Item{}, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType([]*types.Item{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return([]*types.Item{}, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)
		itemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.Item{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleItem,
		).Return(nil)
		helper.service.search = indexManager

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := &types.ItemUpdateInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("without input attached to context", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with no such item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error retrieving item from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error updating item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)
		itemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.Item{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error updating search index", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeItemUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)
		itemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.Item{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleItem,
		).Return(errors.New("blah"))
		helper.service.search = indexManager

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager)
	})
}

func TestItemsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Delete",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return(nil)
		helper.service.search = indexManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, unitCounter)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such item in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error saving as archived", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Delete",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return(errors.New("blah"))
		helper.service.search = indexManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, unitCounter)
	})
}

func TestAccountsService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetAuditLogEntriesForItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType([]*types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetAuditLogEntriesForItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetAuditLogEntriesForItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}
