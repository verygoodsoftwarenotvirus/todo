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
	suite.Run(t, new(auditServiceHTTPRoutesTestSuite))
}

func (s *auditServiceHTTPRoutesTestSuite) TestAuditLogEntriesService_ListHandler() {
	t := s.T()

	exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return(exampleAuditLogEntryList, nil)
	s.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntryList{}))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (s *auditServiceHTTPRoutesTestSuite) TestAuditLogEntriesService_ListHandler_WithNoRowsReturned() {
	t := s.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AuditLogEntryList)(nil), sql.ErrNoRows)
	s.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntryList{}))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (s *auditServiceHTTPRoutesTestSuite) TestAuditLogEntriesService_ListHandler_WithErrorFetchingEntriesFromDatabase() {
	t := s.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AuditLogEntryList)(nil), errors.New("blah"))
	s.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ListHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (s *auditServiceHTTPRoutesTestSuite) TestAuditLogEntriesService_ReadHandler() {
	t := s.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), s.exampleAuditLogEntry.ID).Return(s.exampleAuditLogEntry, nil)
	s.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntry{}))
	s.service.encoderDecoder = ed

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (s *auditServiceHTTPRoutesTestSuite) TestAuditLogEntriesService_ReadHandler_WithNoMatchInDatabase() {
	t := s.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), s.exampleAuditLogEntry.ID).Return((*types.AuditLogEntry)(nil), sql.ErrNoRows)
	s.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusNotFound, s.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}

func (s *auditServiceHTTPRoutesTestSuite) TestAuditLogEntriesService_ReadHandler_WithErrorReadingFromDatabase() {
	t := s.T()

	auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
	auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), s.exampleAuditLogEntry.ID).Return((*types.AuditLogEntry)(nil), errors.New("blah"))
	s.service.auditLog = auditLogEntryManager

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
	s.service.encoderDecoder = ed

	s.service.ReadHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)
	mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
}
