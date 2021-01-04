package accounts

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountsService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccountList := fakes.BuildFakeAccountList()

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccounts", mock.Anything, exampleUser.ID, mock.AnythingOfType("*types.QueryFilter")).Return(exampleAccountList, nil)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountList"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccounts", mock.Anything, exampleUser.ID, mock.AnythingOfType("*types.QueryFilter")).Return((*types.AccountList)(nil), sql.ErrNoRows)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountList"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching accounts from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccounts", mock.Anything, exampleUser.ID, mock.AnythingOfType("*types.QueryFilter")).Return((*types.AccountList)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("CreateAccount", mock.Anything, mock.AnythingOfType("*types.AccountCreationInput")).Return(exampleAccount, nil)
		s.accountDataManager = accountDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.accountCounter = mc

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogAccountCreationEvent", mock.Anything, exampleAccount)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*types.Account"), http.StatusCreated)
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

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, mc, ed)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeInvalidInputResponse", mock.Anything, mock.Anything)
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

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error creating account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("CreateAccount", mock.Anything, mock.AnythingOfType("*types.AccountCreationInput")).Return((*types.Account)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
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

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("AccountExists", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(true, nil)
		s.accountDataManager = accountDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager)
	})

	T.Run("with no such account in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("AccountExists", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(false, sql.ErrNoRows)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
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

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching account from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("AccountExists", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(false, errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
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

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(exampleAccount, nil)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.Account"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with no such account in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching account from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(exampleAccount, nil)
		accountDataManager.On("UpdateAccount", mock.Anything, mock.AnythingOfType("*types.Account")).Return(nil)
		s.accountDataManager = accountDataManager

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogAccountUpdateEvent", mock.Anything, exampleUser.ID, exampleAccount.ID, mock.AnythingOfType("[]types.FieldChangeSummary"))
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.Account"))
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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, s.accountDataManager, s.auditLog, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeInvalidInputResponse", mock.Anything, mock.Anything)
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

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with no rows fetching account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error updating account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(exampleAccount, nil)
		accountDataManager.On("UpdateAccount", mock.Anything, mock.AnythingOfType("*types.Account")).Return(errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(nil)
		s.accountDataManager = accountDataManager

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogAccountArchiveEvent", mock.Anything, exampleUser.ID, exampleAccount.ID)
		s.auditLog = auditLog

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.accountCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, mc)
	})

	T.Run("with no account in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(sql.ErrNoRows)
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
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

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(errors.New("blah"))
		s.accountDataManager = accountDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
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

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.Anything, exampleAccount.ID, exampleUser.ID).Return(nil)
		s.accountDataManager = accountDataManager

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogAccountArchiveEvent", mock.Anything, exampleUser.ID, exampleAccount.ID)
		s.auditLog = auditLog

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.accountCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, mc)
	})
}
