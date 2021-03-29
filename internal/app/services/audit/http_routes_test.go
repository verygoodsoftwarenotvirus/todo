package audit

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountsServiceHTTPRoutes(t *testing.T) {
	suite.Run(t, new(auditServiceHTTPRoutesTestHelper))
}

func (helper *auditServiceHTTPRoutesTestHelper) TestAuditLogEntriesService_ListHandler() {
	t := helper.T()

	exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return(exampleAuditLogEntryList, nil)
	helper.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntryList{}))
	helper.service.encoderDecoder = ed

	helper.service.ListHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (helper *auditServiceHTTPRoutesTestHelper) TestAuditLogEntriesService_ListHandler_WithNoRowsReturned() {
	t := helper.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AuditLogEntryList)(nil), sql.ErrNoRows)
	helper.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntryList{}))
	helper.service.encoderDecoder = ed

	helper.service.ListHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (helper *auditServiceHTTPRoutesTestHelper) TestAuditLogEntriesService_ListHandler_WithErrorFetchingEntriesFromDatabase() {
	t := helper.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AuditLogEntryList)(nil), errors.New("blah"))
	helper.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ListHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (helper *auditServiceHTTPRoutesTestHelper) TestAuditLogEntriesService_ReadHandler() {
	t := helper.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAuditLogEntry.ID).Return(helper.exampleAuditLogEntry, nil)
	helper.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntry{}))
	helper.service.encoderDecoder = ed

	helper.service.ReadHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (helper *auditServiceHTTPRoutesTestHelper) TestAuditLogEntriesService_ReadHandler_WithNoMatchInDatabase() {
	t := helper.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAuditLogEntry.ID).Return((*types.AuditLogEntry)(nil), sql.ErrNoRows)
	helper.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ReadHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusNotFound, helper.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (helper *auditServiceHTTPRoutesTestHelper) TestAuditLogEntriesService_ReadHandler_WithErrorReadingFromDatabase() {
	t := helper.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAuditLogEntry.ID).Return((*types.AuditLogEntry)(nil), errors.New("blah"))
	helper.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	helper.service.encoderDecoder = ed

	helper.service.ReadHandler(helper.res, helper.req)

	assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}
