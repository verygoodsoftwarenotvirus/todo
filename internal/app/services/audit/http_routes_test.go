package audit

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuditLogEntriesService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return(exampleAuditLogEntryList, nil)
		s.auditLog = auditLogEntryManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntryList{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AuditLogEntryList)(nil), sql.ErrNoRows)
		s.auditLog = auditLogEntryManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntryList{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with error fetching entries from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntries", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AuditLogEntryList)(nil), errors.New("blah"))
		s.auditLog = auditLogEntryManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})
}

func TestAuditLogEntriesService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		s.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
			return exampleAuditLogEntry.ID
		}

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), exampleAuditLogEntry.ID).Return(exampleAuditLogEntry, nil)
		s.auditLog = auditLogEntryManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AuditLogEntry{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with no such entry in database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		s.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
			return exampleAuditLogEntry.ID
		}

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), exampleAuditLogEntry.ID).Return((*types.AuditLogEntry)(nil), sql.ErrNoRows)
		s.auditLog = auditLogEntryManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with error fetching entry from database", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		s.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
			return exampleAuditLogEntry.ID
		}

		auditLogEntryManager := &mocktypes.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntry", mock.MatchedBy(testutil.ContextMatcher), exampleAuditLogEntry.ID).Return((*types.AuditLogEntry)(nil), errors.New("blah"))
		s.auditLog = auditLogEntryManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})
}
