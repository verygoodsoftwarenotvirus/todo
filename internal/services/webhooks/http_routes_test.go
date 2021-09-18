package webhooks

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhooksService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeWebhookCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutils.ContextMatcher).Return()
		helper.service.webhookCounter = unitCounter

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			mock.IsType(&types.WebhookCreationInput{}),
		).Return(helper.exampleWebhook, nil)
		helper.service.webhookDataManager = wd

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, unitCounter, wd)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeWebhookCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutils.ContextMatcher,
			testutils.HTTPRequestMatcher,
			mock.IsType(&types.WebhookCreationInput{}),
		).Return(errors.New("blah"))

		encoderDecoder.On(
			"EncodeErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType(""),
			http.StatusBadRequest,
		).Return(nil)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with invalid content attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := &types.WebhookCreationInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutils.ContextMatcher).Return()
		helper.service.webhookCounter = unitCounter

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error creating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleCreationInput := fakes.BuildFakeWebhookCreationInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleCreationInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			testutils.ContextMatcher,
			mock.IsType(&types.WebhookCreationInput{}),
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd)
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
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with no such webhook in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleUpdateInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			testutils.ContextMatcher,
			mock.IsType(&types.Webhook{}),
		).Return(nil)
		helper.service.webhookDataManager = wd

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutils.ContextMatcher,
			testutils.HTTPRequestMatcher,
			mock.IsType(&types.WebhookUpdateInput{}),
		).Return(errors.New("blah"))

		encoderDecoder.On(
			"EncodeErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
			mock.IsType(""),
			http.StatusBadRequest,
		).Return(nil)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with invalid content attached to request", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := &types.WebhookUpdateInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleUpdateInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with no rows fetching webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleUpdateInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleUpdateInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with no rows updating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleUpdateInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			testutils.ContextMatcher,
			mock.IsType(&types.Webhook{}),
		).Return(sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), encoding.ContentTypeJSON)

		exampleUpdateInput := fakes.BuildFakeWebhookUpdateInput()
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, exampleUpdateInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			testutils.ContextMatcher,
			mock.IsType(&types.Webhook{}),
		).Return(errors.New("blah"))
		helper.service.webhookDataManager = wd

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd)
	})
}

func TestWebhooksService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutils.ContextMatcher).Return()
		helper.service.webhookCounter = unitCounter

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(nil)
		helper.service.webhookDataManager = wd

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, unitCounter, wd)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with no webhook in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
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
			testutils.ContextMatcher,
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutils.ContextMatcher,
			testutils.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}
