package audit

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuditLogEntriesService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*models.SessionInfo, error) {
		return &models.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: exampleUser.IsAdmin}, nil
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()

		auditLogEntryManager := &mockmodels.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntries", mock.Anything, mock.AnythingOfType("*models.QueryFilter")).Return(exampleAuditLogEntryList, nil)
		s.auditLog = auditLogEntryManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.AuditLogEntryList"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService()

		s.sessionInfoFetcher = sessionInfoFetcher

		auditLogEntryManager := &mockmodels.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntries", mock.Anything, mock.AnythingOfType("*models.QueryFilter")).Return((*models.AuditLogEntryList)(nil), sql.ErrNoRows)
		s.auditLog = auditLogEntryManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.AuditLogEntryList"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with error fetching entries from database", func(t *testing.T) {
		s := buildTestService()

		s.sessionInfoFetcher = sessionInfoFetcher

		auditLogEntryManager := &mockmodels.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntries", mock.Anything, mock.AnythingOfType("*models.QueryFilter")).Return((*models.AuditLogEntryList)(nil), errors.New("blah"))
		s.auditLog = auditLogEntryManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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

	exampleUser := fakemodels.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*models.SessionInfo, error) {
		return &models.SessionInfo{UserID: exampleUser.ID, UserIsAdmin: exampleUser.IsAdmin}, nil
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		s.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
			return exampleAuditLogEntry.ID
		}

		auditLogEntryManager := &mockmodels.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntry", mock.Anything, exampleAuditLogEntry.ID).Return(exampleAuditLogEntry, nil)
		s.auditLog = auditLogEntryManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.AuditLogEntry"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, auditLogEntryManager, ed)
	})

	T.Run("with no such entry in database", func(t *testing.T) {
		s := buildTestService()

		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		s.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
			return exampleAuditLogEntry.ID
		}

		auditLogEntryManager := &mockmodels.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntry", mock.Anything, exampleAuditLogEntry.ID).Return((*models.AuditLogEntry)(nil), sql.ErrNoRows)
		s.auditLog = auditLogEntryManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		s.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
			return exampleAuditLogEntry.ID
		}

		auditLogEntryManager := &mockmodels.AuditLogEntryDataManager{}
		auditLogEntryManager.On("GetAuditLogEntry", mock.Anything, exampleAuditLogEntry.ID).Return((*models.AuditLogEntry)(nil), errors.New("blah"))
		s.auditLog = auditLogEntryManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
