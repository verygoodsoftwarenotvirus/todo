package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/alexedwards/scs/v2/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_StatusHandler(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{
				UserID:           exampleUser.ID,
				UserIsAdmin:      true,
				AdminPermissions: testutil.BuildMaxAdminPerms(),
			}, nil
		}

		exampleUserToBeBanned := fakes.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUserToBeBanned.ID
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserBanEvent", mock.Anything, exampleUser.ID, exampleUserToBeBanned.ID).Return()
		s.auditLog = auditLog

		udb := &mockmodels.AdminUserDataManager{}
		udb.On(
			"BanUserAccount",
			mock.Anything,
			exampleUserToBeBanned.ID,
		).Return(nil)
		s.userDB = udb

		s.BanHandler(res, req)
		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, udb, auditLog)
	})

	T.Run("with error fetching session", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.BanHandler(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{
				UserID:           exampleUser.ID,
				UserIsAdmin:      false,
				AdminPermissions: testutil.BuildNoAdminPerms(),
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.BanHandler(res, req)
		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("with admin user that does not have permission to ban users", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{
				UserID:           exampleUser.ID,
				UserIsAdmin:      true,
				AdminPermissions: testutil.BuildNoAdminPerms(),
			}, nil
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.BanHandler(res, req)
		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("returns 404 when user doesn't exist", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{
				UserID:           exampleUser.ID,
				UserIsAdmin:      true,
				AdminPermissions: testutil.BuildMaxAdminPerms(),
			}, nil
		}

		exampleUserToBeBanned := fakes.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUserToBeBanned.ID
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		udb := &mockmodels.AdminUserDataManager{}
		udb.On(
			"BanUserAccount",
			mock.Anything,
			exampleUserToBeBanned.ID,
		).Return(sql.ErrNoRows)
		s.userDB = udb

		s.BanHandler(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("with error banning user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		ctx, err := s.sessionManager.Load(context.Background(), "")
		require.NoError(t, err)

		exampleUser := fakes.BuildFakeUser()
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{
				UserID:           exampleUser.ID,
				UserIsAdmin:      true,
				AdminPermissions: testutil.BuildMaxAdminPerms(),
			}, nil
		}

		exampleUserToBeBanned := fakes.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUserToBeBanned.ID
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		udb := &mockmodels.AdminUserDataManager{}
		udb.On(
			"BanUserAccount",
			mock.Anything,
			exampleUserToBeBanned.ID,
		).Return(errors.New("blah"))
		s.userDB = udb

		s.BanHandler(res, req)
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
		s.sessionInfoFetcher = func(*http.Request) (*types.SessionInfo, error) {
			return &types.SessionInfo{
				UserID:           exampleUser.ID,
				UserIsAdmin:      true,
				AdminPermissions: testutil.BuildMaxAdminPerms(),
			}, nil
		}

		exampleUserToBeBanned := fakes.BuildFakeUser()
		s.userIDFetcher = func(req *http.Request) uint64 {
			return exampleUserToBeBanned.ID
		}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://blah.com", nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		auditLog := &mockmodels.AuditLogDataManager{}
		auditLog.On("LogUserBanEvent", mock.Anything, exampleUser.ID, exampleUserToBeBanned.ID).Return()
		s.auditLog = auditLog

		udb := &mockmodels.AdminUserDataManager{}
		udb.On(
			"BanUserAccount",
			mock.Anything,
			exampleUserToBeBanned.ID,
		).Return(nil)
		s.userDB = udb

		s.BanHandler(res, req)
		assert.Equal(t, http.StatusAccepted, res.Code)

		mock.AssertExpectationsForObjects(t, udb)
	})
}
