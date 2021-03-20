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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountsService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		exampleAccountList := fakes.BuildFakeAccountList()

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return(exampleAccountList, nil)
		s.accountDataManager = accountDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountList{}))
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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return((*types.AccountList)(nil), sql.ErrNoRows)
		s.accountDataManager = accountDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountList{}))
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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching accounts from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccounts", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, mock.IsType(&types.QueryFilter{})).Return((*types.AccountList)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{})).Return(exampleAccount, nil)
		s.accountDataManager = accountDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
		s.accountCounter = mc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}), http.StatusCreated)
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

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("CreateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountCreationInput{})).Return((*types.Account)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

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

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return(exampleAccount, nil)
		s.accountDataManager = accountDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}))
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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with no such account in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
		s.accountDataManager = accountDataManager

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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching account from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

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

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return(exampleAccount, nil)
		accountDataManager.On("UpdateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Account{})).Return(nil)
		s.accountDataManager = accountDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Account{}))
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

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, s.accountDataManager, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), sql.ErrNoRows)
		s.accountDataManager = accountDataManager

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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error fetching account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return((*types.Account)(nil), errors.New("blah"))
		s.accountDataManager = accountDataManager

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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error updating account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(exampleAccount)

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("GetAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return(exampleAccount, nil)
		accountDataManager.On("UpdateAccount", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Account{})).Return(errors.New("blah"))
		s.accountDataManager = accountDataManager

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

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}

func TestAccountsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return(nil)
		s.accountDataManager = accountDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
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

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return(sql.ErrNoRows)
		s.accountDataManager = accountDataManager

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

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()

		s.accountIDFetcher = func(req *http.Request) uint64 {
			return exampleAccount.ID
		}

		s.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
			reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
			require.NoError(t, err)
			return reqCtx, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On("ArchiveAccount", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleUser.ID).Return(errors.New("blah"))
		s.accountDataManager = accountDataManager

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

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, ed)
	})
}
