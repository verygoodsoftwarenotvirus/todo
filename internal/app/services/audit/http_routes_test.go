package audit

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuditLogEntriesService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On(
			"GetAuditLogEntries",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAuditLogEntryList, nil)
		helper.service.auditLog = auditLogEntryManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.AuditLogEntryList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, encoderDecoder)
	})

	T.Run("with no results returned from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On(
			"GetAuditLogEntries",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AuditLogEntryList)(nil), sql.ErrNoRows)
		helper.service.auditLog = auditLogEntryManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.AuditLogEntryList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, encoderDecoder)
	})
	T.Run("with error reading from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On(
			"GetAuditLogEntries",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AuditLogEntryList)(nil), errors.New("blah"))
		helper.service.auditLog = auditLogEntryManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, encoderDecoder)
	})
}

func TestAuditLogEntriesService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On(
			"GetAuditLogEntry",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAuditLogEntry.ID,
		).Return(helper.exampleAuditLogEntry, nil)
		helper.service.auditLog = auditLogEntryManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
			mock.IsType(&types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, encoderDecoder)
	})

	T.Run("with no audit log entries returned from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On(
			"GetAuditLogEntry",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAuditLogEntry.ID,
		).Return((*types.AuditLogEntry)(nil), sql.ErrNoRows)
		helper.service.auditLog = auditLogEntryManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, encoderDecoder)
	})

	T.Run("with error reading from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On(
			"GetAuditLogEntry",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleAuditLogEntry.ID,
		).Return((*types.AuditLogEntry)(nil), errors.New("blah"))
		helper.service.auditLog = auditLogEntryManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.ResponseWriterMatcher),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, encoderDecoder)
	})
}
