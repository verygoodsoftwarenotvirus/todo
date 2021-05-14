package admin

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/alexedwards/scs/v2/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminService_UserAccountStatusChangeHandler_BanningAccounts(T *testing.T) {
	T.Parallel()

	T.Run("with no input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		var err error
		helper.req, err = http.NewRequest(http.MethodGet, "/blah", nil)
		require.NoError(t, err)

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.exampleInput.NewReputation = types.BannedUserReputation

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})

	T.Run("banning users", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleInput.NewReputation = types.BannedUserReputation

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(nil)
		helper.service.userDB = userDataManager

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On(
			"LogUserBanEvent",
			testutil.ContextMatcher,
			helper.exampleUser.ID, helper.exampleInput.TargetUserID, helper.exampleInput.Reason,
		).Return()
		helper.service.auditLog = auditLog

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager, auditLog)
	})

	T.Run("terminating users", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleInput.NewReputation = types.TerminatedUserReputation

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(nil)
		helper.service.userDB = userDataManager

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On(
			"LogAccountTerminationEvent",
			testutil.ContextMatcher,
			helper.exampleUser.ID, helper.exampleInput.TargetUserID, helper.exampleInput.Reason,
		).Return()
		helper.service.auditLog = auditLog

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager, auditLog)
	})

	T.Run("back in good standing", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(nil)
		helper.service.userDB = userDataManager

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with inadequate admin user attempting to ban", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleInput.NewReputation = types.BannedUserReputation

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			scd := &types.SessionContextData{
				Requester: types.RequesterInfo{
					ServiceAdminPermission: permissions.ServiceAdminPermission(1),
				},
			}

			return scd, nil
		}

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)

		mock.AssertExpectationsForObjects(t)
	})

	T.Run("with inadequate admin user attempting to terminate accounts", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleInput.NewReputation = types.TerminatedUserReputation

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			scd := &types.SessionContextData{
				Requester: types.RequesterInfo{
					ServiceAdminPermission: permissions.ServiceAdminPermission(1),
				},
			}

			return scd, nil
		}

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)

		mock.AssertExpectationsForObjects(t)
	})

	T.Run("with admin that has inadequate permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.neuterAdminUser()

		helper.exampleInput.NewReputation = types.BannedUserReputation

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with no such user in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleInput.NewReputation = types.BannedUserReputation

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(sql.ErrNoRows)
		helper.service.userDB = userDataManager

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with error writing new reputation to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleInput.NewReputation = types.BannedUserReputation

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(errors.New("blah"))
		helper.service.userDB = userDataManager

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with error destroying session", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		mockHandler := &mockstore.MockStore{}
		mockHandler.ExpectDelete("", errors.New("blah"))
		helper.service.sessionManager.Store = mockHandler

		helper.exampleInput.NewReputation = types.BannedUserReputation

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On(
			"LogUserBanEvent",
			testutil.ContextMatcher,
			helper.exampleUser.ID, helper.exampleInput.TargetUserID, helper.exampleInput.Reason,
		).Return()
		helper.service.auditLog = auditLog

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(nil)
		helper.service.userDB = userDataManager

		helper.service.UserAccountStatusChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, userDataManager)
	})
}
