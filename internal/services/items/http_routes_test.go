package items

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			testutils.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutils.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleItem,
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

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
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
			testutils.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutils.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleItem,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType(&types.Item{}),
		)
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
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		)
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
			testutils.ContextMatcher,
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
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(false, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(false, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return(exampleItemList.Items, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return([]uint64{}, errors.New("blah"))
		helper.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return([]*types.Item{}, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return([]*types.Item{}, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)

		itemDataManager.On(
			"UpdateItem",
			testutils.ContextMatcher,
			mock.IsType(&types.Item{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleItem,
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
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
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
			testutils.ContextMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)

		itemDataManager.On(
			"UpdateItem",
			testutils.ContextMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)

		itemDataManager.On(
			"UpdateItem",
			testutils.ContextMatcher,
			mock.IsType(&types.Item{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleItem,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Delete",
			testutils.ContextMatcher,
			helper.exampleItem.ID,
		).Return(nil)
		helper.service.search = indexManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutils.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, unitCounter, indexManager)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Delete",
			testutils.ContextMatcher,
			helper.exampleItem.ID,
		).Return(errors.New("blah"))
		helper.service.search = indexManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutils.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, unitCounter, indexManager)
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleItem.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}
