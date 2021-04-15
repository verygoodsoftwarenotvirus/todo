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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebhooksService_List(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.WebhookList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
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
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_Create(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", mock.MatchedBy(testutil.ContextMatcher)).Return()
		helper.service.webhookCounter = unitCounter

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.WebhookCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleWebhook, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.Webhook{}), http.StatusCreated)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, unitCounter, wd, encoderDecoder)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.req = testutil.BuildTestRequest(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
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
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.WebhookCreationInput{}),
			helper.exampleUser.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_Read(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.Webhook{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with no such webhook in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
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
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.Webhook{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.Webhook{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)
		helper.req = testutil.BuildTestRequest(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
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
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
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
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
		).Return(helper.exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.Webhook{}),
			helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}

func TestWebhooksService_Archive(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
		helper.service.webhookCounter = unitCounter

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.webhookDataManager = wd

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, unitCounter, wd)
	})

	T.Run("with no webhook in database", func(t *testing.T) {
		t.Parallel()

		helper := newTestHelper(t)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
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
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleWebhook.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.webhookDataManager = wd

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, wd, encoderDecoder)
	})
}
