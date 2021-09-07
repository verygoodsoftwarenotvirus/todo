package authentication

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildArbitraryPASETO(t *testing.T, helper *authServiceHTTPRoutesTestHelper, issueTime time.Time, lifetime time.Duration, pasetoData string) *types.PASETOResponse {
	t.Helper()

	jsonToken := paseto.JSONToken{
		Audience:   helper.exampleAPIClient.BelongsToUser,
		Subject:    helper.exampleAPIClient.BelongsToUser,
		Jti:        uuid.NewString(),
		Issuer:     helper.service.config.PASETO.Issuer,
		IssuedAt:   issueTime,
		NotBefore:  issueTime,
		Expiration: issueTime.Add(lifetime),
	}

	jsonToken.Set(pasetoDataKey, pasetoData)

	// Encrypt data
	token, err := paseto.NewV2().Encrypt(helper.service.config.PASETO.LocalModeKey, jsonToken, "")
	require.NoError(t, err)

	return &types.PASETOResponse{
		Token:     token,
		ExpiresAt: jsonToken.Expiration.String(),
	}
}

func TestService_fetchSessionContextDataFromPASETO(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes, err := helper.service.buildPASETOResponse(helper.ctx, helper.sessionCtxData, helper.exampleAPIClient)
		require.NoError(t, err)

		helper.req.Header.Set(pasetoAuthorizationHeaderKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid PASETO", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.Header.Set(pasetoAuthorizationHeaderKey, "blah")

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with expired token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes := buildArbitraryPASETO(t, helper, time.Now().Add(-24*time.Hour), time.Minute, base64.RawURLEncoding.EncodeToString(helper.sessionCtxData.ToBytes()))

		helper.req.Header.Set(pasetoAuthorizationHeaderKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid base64 encoding", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes := buildArbitraryPASETO(t, helper, time.Now(), time.Hour, `       \\\\\\\\\\\\               lololo`)

		helper.req.Header.Set(pasetoAuthorizationHeaderKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid GOB string", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes := buildArbitraryPASETO(t, helper, time.Now(), time.Hour, base64.RawURLEncoding.EncodeToString([]byte("blah")))

		helper.req.Header.Set(pasetoAuthorizationHeaderKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with missing token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestAuthenticationService_CookieAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountUserMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountUserMembershipDataManager.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.sessionCtxData, nil)
		helper.service.accountMembershipManager = accountUserMembershipDataManager

		mockHandler := &testutils.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		_, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		helper.service.CookieRequirementMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})
}

func TestAuthenticationService_UserAttributionMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		mockAccountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		mockAccountMembershipManager.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sessionCtxData, nil)
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		h := &testutils.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockAccountMembershipManager, h)
	})

	T.Run("with error building session context data for user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		mockAccountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		mockAccountMembershipManager.On(
			"BuildSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.SessionContextData)(nil), errors.New("blah"))
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		mh := &testutils.MockHTTPHandler{}
		helper.service.UserAttributionMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockAccountMembershipManager, mh)
	})

	T.Run("with PASETO", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes, err := helper.service.buildPASETOResponse(helper.ctx, helper.sessionCtxData, helper.exampleAPIClient)
		require.NoError(t, err)

		helper.req.Header.Set(pasetoAuthorizationHeaderKey, tokenRes.Token)

		h := &testutils.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with PASETO and issue parsing token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.req.Header.Set(pasetoAuthorizationHeaderKey, "blah")

		h := &testutils.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})
}

func TestAuthenticationService_AuthorizationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		mockUserDataManager := &mocktypes.UserDataManager{}
		mockUserDataManager.On(
			"GetSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sessionCtxData, nil)
		helper.service.userDataManager = mockUserDataManager

		h := &testutils.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.SessionContextDataKey, sessionCtxData))

		helper.service.AuthorizationMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceAccountStatus = types.BannedUserAccountStatus
		helper.setContextFetcher(t)

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		mockUserDataManager := &mocktypes.UserDataManager{}
		mockUserDataManager.On(
			"GetSessionContextDataForUser",
			testutils.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sessionCtxData, nil)
		helper.service.userDataManager = mockUserDataManager

		h := &testutils.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.SessionContextDataKey, sessionCtxData))

		mh := &testutils.MockHTTPHandler{}
		helper.service.AuthorizationMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with missing session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, nil
		}

		mh := &testutils.MockHTTPHandler{}
		helper.service.AuthorizationMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("without authorization for account", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		sessionCtxData.AccountPermissions = map[string]authorization.AccountRolePermissionsChecker{}
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return sessionCtxData, nil
		}

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.SessionContextDataKey, sessionCtxData))

		helper.service.AuthorizationMiddleware(&testutils.MockHTTPHandler{}).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})
}

func TestAuthenticationService_AdminMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceRoles = []string{authorization.ServiceAdminRole.String()}
		helper.setContextFetcher(t)

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.SessionContextDataKey, sessionCtxData))

		mockHandler := &testutils.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutils.HTTPResponseWriterMatcher,
			testutils.HTTPRequestMatcher,
		).Return()

		helper.service.ServiceAdminMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceRoles = []string{authorization.ServiceAdminRole.String()}
		helper.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.SessionContextDataKey, sessionCtxData))

		mockHandler := &testutils.MockHTTPHandler{}
		helper.service.ServiceAdminMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData := &types.SessionContextData{
			Requester: types.RequesterInfo{
				UserID:                helper.exampleUser.ID,
				Reputation:            helper.exampleUser.ServiceAccountStatus,
				ReputationExplanation: helper.exampleUser.ReputationExplanation,
				ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
			},
			ActiveAccountID:    helper.exampleAccount.ID,
			AccountPermissions: helper.examplePermCheckers,
		}

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.SessionContextDataKey, sessionCtxData))

		mockHandler := &testutils.MockHTTPHandler{}
		helper.service.ServiceAdminMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})
}
