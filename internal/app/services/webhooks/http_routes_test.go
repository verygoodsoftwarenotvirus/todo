package webhooks

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebhooksService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.webhookCounter = unitCounter

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			testutil.ContextMatcher,
			mock.IsType(&types.WebhookCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleWebhook, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Webhook{}), http.StatusCreated)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusCreated, helper.res.Code)
		mock.AssertExpectationsForObjects(t, unitCounter, wd, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.req = testutil.BuildTestRequest(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error creating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			testutil.ContextMatcher,
			mock.IsType(&types.WebhookCreationInput{}),
			helper.exampleUser.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.WebhookList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.WebhookList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error fetching webhooks from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Webhook{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with no such webhook in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error fetching webhook from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			testutil.ContextMatcher,
			mock.IsType(&types.Webhook{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Webhook{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.req = testutil.BuildTestRequest(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no rows fetching webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with no rows updating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			testutil.ContextMatcher,
			mock.IsType(&types.Webhook{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			testutil.ContextMatcher,
			mock.IsType(&types.Webhook{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.webhookCounter = unitCounter

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.webhookDataManager = wd

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNoContent, helper.res.Code)
		mock.AssertExpectationsForObjects(t, unitCounter, wd)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with no webhook in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		webhookDataManager := &mocktypes.WebhookDataManager{}
		webhookDataManager.On(
			"GetAuditLogEntriesForWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.webhookDataManager = webhookDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, webhookDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
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

		helper := newTestHelper(t)

		webhookDataManager := &mocktypes.WebhookDataManager{}
		webhookDataManager.On(
			"GetAuditLogEntriesForWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = webhookDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, webhookDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		webhookDataManager := &mocktypes.WebhookDataManager{}
		webhookDataManager.On(
			"GetAuditLogEntriesForWebhook",
			testutil.ContextMatcher,
			helper.exampleWebhook.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.webhookDataManager = webhookDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, webhookDataManager, encoderDecoder)
	})
}
