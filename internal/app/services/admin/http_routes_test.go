package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/alexedwards/scs/v2/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_UserAccountStatusChangeHandler(T *testing.T) {
	T.Parallel()

	T.Run("banning users happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildMaxServiceAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		udb := &mocktypes.AdminUserDataManager{}
		udb.On(
			"UpdateUserAccountStatus",
			mock.Anything,
			exampleInput.TargetAccountID,
			*exampleInput,
		).Return(nil)
		s.userDB = udb

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, exampleInput.TargetAccountID, exampleInput.Reason).Return()
		s.auditLog = auditLog

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("terminating accounts happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildMaxServiceAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.TerminatedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		udb := &mocktypes.AdminUserDataManager{}
		udb.On(
			"UpdateUserAccountStatus",
			mock.Anything,
			exampleInput.TargetAccountID,
			*exampleInput,
		).Return(nil)
		s.userDB = udb

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogAccountTerminationEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, exampleInput.TargetAccountID, exampleInput.Reason).Return()
		s.auditLog = auditLog

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildMaxServiceAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.UserAccountStatusChangeHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error fetching session", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildNoAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with admin user that does not have the right permissions", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildNoAdminPerms() + 1,
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("returns 404 when user does not exist", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildMaxServiceAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		udb := &mocktypes.AdminUserDataManager{}
		udb.On(
			"UpdateUserAccountStatus",
			mock.Anything,
			exampleInput.TargetAccountID,
			*exampleInput,
		).Return(sql.ErrNoRows)
		s.userDB = udb

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error banning user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildMaxServiceAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		udb := &mocktypes.AdminUserDataManager{}
		udb.On(
			"UpdateUserAccountStatus",
			mock.Anything,
			exampleInput.TargetAccountID,
			*exampleInput,
		).Return(errors.New("blah"))
		s.userDB = udb

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error destroying session", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		ms := &mockstore.MockStore{}
		ms.ExpectDelete("", errors.New("blah"))
		s.sessionManager.Store = ms

		exampleUser := fakes.BuildFakeUser()
		s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return &types.RequestContext{
				User: types.UserRequestContext{
					ID:                      exampleUser.ID,
					ServiceAdminPermissions: testutil.BuildMaxServiceAdminPerms(),
				},
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		exampleInput.NewReputation = types.BannedAccountStatus

		req = req.WithContext(context.WithValue(req.Context(), accountStatusUpdateMiddlewareCtxKey, exampleInput))

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogUserBanEvent", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID, exampleInput.TargetAccountID, exampleInput.Reason).Return()
		s.auditLog = auditLog

		udb := &mocktypes.AdminUserDataManager{}
		udb.On(
			"UpdateUserAccountStatus",
			mock.Anything,
			exampleInput.TargetAccountID,
			*exampleInput,
		).Return(nil)
		s.userDB = udb

		s.UserAccountStatusChangeHandler(res, req)
		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})
}
