package admin

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/alexedwards/scs/v2/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminService_UserAccountStatusChangeHandler(T *testing.T) {
	T.Parallel()

	T.Run("banning users", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

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

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager, auditLog)
	})

	T.Run("terminating users", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.TerminatedUserReputation
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

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

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager, auditLog)
	})

	T.Run("back in good standing", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.GoodStandingAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(nil)
		helper.service.userDB = userDataManager

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
	})

	T.Run("with no input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UserReputationChangeHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput = &types.UserReputationUpdateInput{}
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with inadequate admin user attempting to ban", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			scd := &types.SessionContextData{
				Requester: types.RequesterInfo{
					ServicePermissions: authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
				},
			}

			return scd, nil
		}

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)

		mock.AssertExpectationsForObjects(t)
	})

	T.Run("with inadequate admin user attempting to terminate accounts", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.TerminatedUserReputation
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			scd := &types.SessionContextData{
				Requester: types.RequesterInfo{
					ServicePermissions: authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
				},
			}

			return scd, nil
		}

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)

		mock.AssertExpectationsForObjects(t)
	})

	T.Run("with admin that has inadequate permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.neuterAdminUser()

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusForbidden, helper.res.Code)
	})

	T.Run("with no such user in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(sql.ErrNoRows)
		helper.service.userDB = userDataManager

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with error writing new reputation to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		userDataManager := &mocktypes.AdminUserDataManager{}
		userDataManager.On(
			"UpdateUserReputation",
			testutil.ContextMatcher,
			helper.exampleInput.TargetUserID,
			helper.exampleInput,
		).Return(errors.New("blah"))
		helper.service.userDB = userDataManager

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager)
	})

	T.Run("with error destroying session", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.exampleInput.NewReputation = types.BannedUserAccountStatus
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockHandler := &mockstore.MockStore{}
		mockHandler.ExpectDelete("", errors.New("blah"))
		helper.service.sessionManager.Store = mockHandler

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On(
			"LogUserBanEvent",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			helper.exampleInput.TargetUserID,
			helper.exampleInput.Reason,
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

		helper.service.UserReputationChangeHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusAccepted, helper.res.Code)

		mock.AssertExpectationsForObjects(t, userDataManager)
	})
}
