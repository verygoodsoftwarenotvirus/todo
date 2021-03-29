package items

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestItemsService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		exampleItemList := fakes.BuildFakeItemList()

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, mock.IsType(&types.QueryFilter{})).Return(exampleItemList, nil)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.ItemList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, mock.IsType(&types.QueryFilter{})).Return((*types.ItemList)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.ItemList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error retrieving items from database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, mock.IsType(&types.QueryFilter{})).Return((*types.ItemList)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_SearchHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList()
		exampleItemIDs := []uint64{}
		for _, x := range exampleItemList.Items {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		helper.req.URL.RawQuery = url.Values{"q": []string{exampleQuery}, "limit": []string{strconv.Itoa(int(exampleLimit))}}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, helper.exampleAccount.ID).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, exampleLimit, exampleItemIDs).Return(exampleItemList.Items, nil)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType([]*types.Item{}))
		helper.service.encoderDecoder = ed

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, ed)
	})

	T.Run("with error conducting search", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		exampleQuery := "whatever"

		helper.req.URL.RawQuery = url.Values{"q": []string{exampleQuery}}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, helper.exampleAccount.ID).Return([]uint64{}, errors.New("blah"))
		helper.service.search = indexManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList()
		exampleItemIDs := []uint64{}
		for _, x := range exampleItemList.Items {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		helper.req.URL.RawQuery = url.Values{"q": []string{exampleQuery}, "limit": []string{strconv.Itoa(int(exampleLimit))}}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, helper.exampleAccount.ID).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, exampleLimit, exampleItemIDs).Return([]*types.Item{}, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType([]*types.Item{}))
		helper.service.encoderDecoder = ed

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, ed)
	})

	T.Run("with error retrieving from database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList()
		exampleItemIDs := []uint64{}
		for _, x := range exampleItemList.Items {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		helper.req.URL.RawQuery = url.Values{"q": []string{exampleQuery}, "limit": []string{strconv.Itoa(int(exampleLimit))}}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, helper.exampleAccount.ID).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccount.ID, exampleLimit, exampleItemIDs).Return([]*types.Item{}, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, ed)
	})
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.ItemCreationInput{})).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
		helper.service.itemCounter = mc

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Index", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleItem).Return(nil)
		helper.service.search = indexManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Item{}), http.StatusCreated)
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, mc, indexManager, ed)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		helper.req = testutil.BuildTestRequest(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error creating item", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.ItemCreationInput{})).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(true, nil)
		helper.service.itemDataManager = itemDataManager

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with no result in the database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(false, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error checking database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(false, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		h := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), h.exampleItem.ID, h.exampleAccount.ID).Return(h.exampleItem, nil)
		h.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Item{}))
		h.service.encoderDecoder = ed

		h.service.ReadHandler(h.res, h.req)

		assert.Equal(t, http.StatusOK, h.res.Code, "expected %d in status response, got %d", http.StatusOK, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with no such item in the database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()
		h := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), h.exampleItem.ID, h.exampleAccount.ID).Return((*types.Item)(nil), errors.New("blah"))
		h.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		h.service.encoderDecoder = ed

		h.service.ReadHandler(h.res, h.req)

		assert.Equal(t, http.StatusInternalServerError, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		h := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), h.exampleItem.ID, h.exampleAccount.ID).Return(h.exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Item{}), mock.IsType([]*types.FieldChangeSummary{})).Return(nil)
		h.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Index", mock.MatchedBy(testutil.ContextMatcher), h.exampleItem.ID, h.exampleItem).Return(nil)
		h.service.search = indexManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Item{}))
		h.service.encoderDecoder = ed

		h.service.UpdateHandler(h.res, h.req)

		assert.Equal(t, http.StatusOK, h.res.Code, "expected %d in status response, got %d", http.StatusOK, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, ed)
	})

	T.Run("without input attached to context", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		helper.req = testutil.BuildTestRequest(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with no such item", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error retrieving item from database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error updating item", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(helper.exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Item{}), mock.IsType([]*types.FieldChangeSummary{})).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		h := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), h.exampleItem.ID, h.exampleAccount.ID).Return(nil)
		h.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Delete", mock.MatchedBy(testutil.ContextMatcher), h.exampleItem.ID).Return(nil)
		h.service.search = indexManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
		h.service.itemCounter = mc

		h.service.ArchiveHandler(h.res, h.req)

		assert.Equal(t, http.StatusNoContent, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, mc)
	})

	T.Run("with no such item in the database", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error saving as archived", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		t.Parallel()
		helper := newTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID, helper.exampleAccount.ID).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Delete", mock.MatchedBy(testutil.ContextMatcher), helper.exampleItem.ID).Return(errors.New("blah"))
		helper.service.search = indexManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
		helper.service.itemCounter = mc

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, mc)
	})
}
