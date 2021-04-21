package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParseLoginInputFromForm(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := &types.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "123456",
		}
		req := &http.Request{}

		form := url.Values{}
		form.Set(usernameFormKey, expected.Username)
		form.Set(passwordFormKey, expected.Password)
		form.Set(totpTokenFormKey, expected.TOTPToken)
		req.Form = form

		actual := parseLoginInputFromForm(req)

		assert.Equal(t, expected, actual)
	})

	T.Run("returns nil for invalid request", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, parseLoginInputFromForm(&http.Request{}))
	})
}

func buildArbitraryPASETO(t *testing.T, helper *authServiceHTTPRoutesTestHelper, issueTime time.Time, lifetime time.Duration, pasetoData string) *types.PASETOResponse {
	t.Helper()

	jsonToken := paseto.JSONToken{
		Audience:   strconv.FormatUint(helper.exampleAPIClient.BelongsToUser, 10),
		Subject:    strconv.FormatUint(helper.exampleAPIClient.BelongsToUser, 10),
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

		helper.req.Header.Set(pasetoAuthorizationKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid PASETO", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.Header.Set(pasetoAuthorizationKey, "blah")

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with expired token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes := buildArbitraryPASETO(t, helper, time.Now().Add(-24*time.Hour), time.Minute, base64.RawURLEncoding.EncodeToString(helper.sessionCtxData.ToBytes()))

		helper.req.Header.Set(pasetoAuthorizationKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid base64 encoding", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes := buildArbitraryPASETO(t, helper, time.Now(), time.Hour, `       \\\\\\\\\\\\               lololo`)

		helper.req.Header.Set(pasetoAuthorizationKey, tokenRes.Token)

		actual, err := helper.service.fetchSessionContextDataFromPASETO(helper.ctx, helper.req)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid GOB string", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes := buildArbitraryPASETO(t, helper, time.Now(), time.Hour, base64.RawURLEncoding.EncodeToString([]byte("fart")))

		helper.req.Header.Set(pasetoAuthorizationKey, tokenRes.Token)

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

func TestAuthService_CookieAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountUserMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountUserMembershipDataManager.On(
			"BuildSessionContextDataForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.sessionCtxData, nil)
		helper.service.accountMembershipManager = accountUserMembershipDataManager

		mockHandler := &testutil.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		_, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		helper.service.CookieRequirementMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})
}

func TestAuthService_UserAttributionMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockAccountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		mockAccountMembershipManager.On(
			"BuildSessionContextDataForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sessionCtxData, nil)
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
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
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.SessionContextData)(nil), errors.New("blah"))
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req, _ = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		mh := &testutil.MockHTTPHandler{}
		helper.service.UserAttributionMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockAccountMembershipManager, mh)
	})

	T.Run("with PASETO", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		tokenRes, err := helper.service.buildPASETOResponse(helper.ctx, helper.sessionCtxData, helper.exampleAPIClient)
		require.NoError(t, err)

		helper.req.Header.Set(pasetoAuthorizationKey, tokenRes.Token)

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with PASETO and issue parsing token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.req.Header.Set(pasetoAuthorizationKey, "blah")

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})
}

func TestAuthService_AuthorizationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockUserDataManager := &mocktypes.UserDataManager{}
		mockUserDataManager.On(
			"GetSessionContextDataForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sessionCtxData, nil)
		helper.service.userDataManager = mockUserDataManager

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.SessionContextDataKey, sessionCtxData))

		helper.service.AuthorizationMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.Reputation = types.BannedUserReputation
		helper.setContextFetcher(t)

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockUserDataManager := &mocktypes.UserDataManager{}
		mockUserDataManager.On(
			"GetSessionContextDataForUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(sessionCtxData, nil)
		helper.service.userDataManager = mockUserDataManager

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.SessionContextDataKey, sessionCtxData))

		mh := &testutil.MockHTTPHandler{}
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

		mh := &testutil.MockHTTPHandler{}
		helper.service.AuthorizationMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("without authorization for account", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		sessionCtxData.AccountPermissionsMap = map[uint64]*types.UserAccountMembershipInfo{}
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return sessionCtxData, nil
		}

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.SessionContextDataKey, sessionCtxData))

		helper.service.AuthorizationMiddleware(&testutil.MockHTTPHandler{}).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})
}

func TestAuthService_PermissionRestrictionMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.
			PermissionRestrictionMiddleware(permissions.CanManageAPIClients)(h).
			ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.
			PermissionRestrictionMiddleware(permissions.CanManageAPIClients)(&testutil.MockHTTPHandler{}).
			ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with admin permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			sessionContextData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, map[uint64]*types.UserAccountMembershipInfo{})
			require.NoError(t, err)

			return sessionContextData, nil
		}

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.
			PermissionRestrictionMiddleware(permissions.CanManageAPIClients)(h).
			ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
	})

	T.Run("without adequate account permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			sessionContextData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, map[uint64]*types.UserAccountMembershipInfo{})
			require.NoError(t, err)

			return sessionContextData, nil
		}

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.
			PermissionRestrictionMiddleware(permissions.CanManageAPIClients)(h).
			ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})

	T.Run("with inadequate account permissions", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			sessionContextData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, map[uint64]*types.UserAccountMembershipInfo{
				helper.exampleAccount.ID: {
					Permissions: permissions.ServiceUserPermission(0),
				},
			})
			require.NoError(t, err)

			return sessionContextData, nil
		}

		h := &testutil.MockHTTPHandler{}
		h.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.
			PermissionRestrictionMiddleware(permissions.CanManageAPIClients)(h).
			ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
	})
}

func TestAuthService_AdminMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
		helper.setContextFetcher(t)

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.SessionContextDataKey, sessionCtxData))

		mockHandler := &testutil.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.AdminMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.SessionContextDataKey, sessionCtxData))

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.AdminMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.SessionContextDataKey, sessionCtxData))

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.AdminMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockHandler)
	})
}

func TestAuthService_ChangeActiveAccountInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeChangeActiveAccountInput()
		jsonBytes, err := json.Marshal(&exampleInput)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		mockHandler := &testutil.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.ChangeActiveAccountInputMiddleware(mockHandler).ServeHTTP(helper.res, req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&types.ChangeActiveAccountInput{AccountID: helper.exampleAccount.ID}))

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.ChangeActiveAccountInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"attached input is invalid",
			http.StatusBadRequest,
		)
		helper.service.encoderDecoder = encoderDecoder

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.ChangeActiveAccountInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mockHandler)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := &types.ChangeActiveAccountInput{}
		jsonBytes, err := json.Marshal(&exampleInput)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.ChangeActiveAccountInputMiddleware(mockHandler).ServeHTTP(helper.res, req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockHandler)
	})
}

func TestAuthService_UserLoginInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(helper.exampleLoginInput))

		var err error
		helper.req, err = http.NewRequest(http.MethodPost, "/login", &b)
		require.NoError(t, err)

		mockHandler := &testutil.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.UserLoginInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(helper.exampleLoginInput))

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.UserLoginInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher, "attached input is invalid", http.StatusBadRequest)
		helper.service.encoderDecoder = encoderDecoder

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.UserLoginInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mockHandler)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := &types.UserLoginInput{}

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		var err error
		helper.req, err = http.NewRequest(http.MethodPost, "/login", &b)
		require.NoError(t, err)

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.UserLoginInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockHandler)
	})
}

func TestAuthService_PASETOCreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(fakes.BuildFakePASETOCreationInput()))

		req, err := http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		helper.req, err = http.NewRequest(http.MethodPost, "/login", &b)
		require.NoError(t, err)

		mockHandler := &testutil.MockHTTPHandler{}
		mockHandler.On(
			"ServeHTTP",
			testutil.ResponseWriterMatcher,
			testutil.RequestMatcher,
		).Return()

		helper.service.PASETOCreationInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, mockHandler)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			testutil.ContextMatcher,
			testutil.RequestMatcher,
			mock.IsType(&types.PASETOCreationInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"attached input is invalid",
			http.StatusBadRequest,
		)
		helper.service.encoderDecoder = encoderDecoder

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.PASETOCreationInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder, mockHandler)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&types.PASETOCreationInput{}))

		req, err := http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		helper.req, err = http.NewRequest(http.MethodPost, "/login", &b)
		require.NoError(t, err)

		mockHandler := &testutil.MockHTTPHandler{}
		helper.service.PASETOCreationInputMiddleware(mockHandler).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})
}
