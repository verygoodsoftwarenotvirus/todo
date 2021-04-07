package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test_parseLoginInputFromForm here

func TestAuthService_CookieAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		md := &mocktypes.UserDataManager{}
		md.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.exampleUser, nil)
		helper.service.userDataManager = md

		aumdm := &mocktypes.AccountUserMembershipDataManager{}
		aumdm.On("BuildRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(helper.reqCtx, nil)
		helper.service.accountMembershipManager = aumdm

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		_, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		helper.service.CookieAuthenticationMiddleware(ms).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, md, ms)
	})

	T.Run("with nil user returned by datastore and without error", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		md := &mocktypes.UserDataManager{}
		md.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return((*types.User)(nil), nil)
		helper.service.userDataManager = md

		_, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		ms := &MockHTTPHandler{}

		helper.service.CookieAuthenticationMiddleware(ms).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, md, ms)
	})

	T.Run("without user attached to request", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		ms := &MockHTTPHandler{}
		helper.service.CookieAuthenticationMiddleware(ms).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, ms)
	})
}

func TestAuthService_UserAttributionMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockAccountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		mockAccountMembershipManager.On("BuildRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(reqCtx, nil)
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockAccountMembershipManager, h)
	})

	T.Run("with PASETO", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockAccountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		mockAccountMembershipManager.On("BuildRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(reqCtx, nil)
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.service.UserAttributionMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockAccountMembershipManager, h)
	})

	T.Run("with error fetching request context for user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		mockAccountMembershipManager := &mocktypes.AccountUserMembershipDataManager{}
		mockAccountMembershipManager.On("BuildRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return((*types.RequestContext)(nil), errors.New("blah"))
		helper.service.accountMembershipManager = mockAccountMembershipManager

		_, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		mh := &testutil.MockHTTPHandler{}
		helper.service.UserAttributionMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockAccountMembershipManager, mh)
	})
}

func TestAuthService_AuthorizationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockUserDataManager := &mocktypes.UserDataManager{}
		mockUserDataManager.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(reqCtx, nil)
		helper.service.userDataManager = mockUserDataManager

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.RequestContextKey, reqCtx))

		helper.service.AuthorizationMiddleware(h).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleUser.Reputation = types.BannedAccountStatus
		helper.setContextFetcher(t)

		reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		mockUserDataManager := &mocktypes.UserDataManager{}
		mockUserDataManager.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), helper.exampleUser.ID).Return(reqCtx, nil)
		helper.service.userDataManager = mockUserDataManager

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.req = helper.req.WithContext(context.WithValue(helper.ctx, types.RequestContextKey, reqCtx))

		mh := &testutil.MockHTTPHandler{}
		helper.service.AuthorizationMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusForbidden, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with missing request context", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
			return nil, nil
		}

		mh := &testutil.MockHTTPHandler{}
		helper.service.AuthorizationMiddleware(mh).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestAuthService_UserLoginInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(helper.exampleLoginInput))

		req, err := http.NewRequestWithContext(helper.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		helper.req, err = http.NewRequest(http.MethodPost, "/login", &b)
		require.NoError(t, err)

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.service.UserLoginInputMiddleware(ms).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(helper.exampleLoginInput))

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.IsType(&types.UserLoginInput{})).Return(errors.New("blah"))
		ed.On("EncodeErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), "attached input is invalid", http.StatusBadRequest)
		helper.service.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		helper.service.UserLoginInputMiddleware(ms).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, ed, ms)
	})

	T.Run("with error decoding request but valid input attached to request form", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		form := url.Values{
			usernameFormKey:  {helper.exampleLoginInput.Username},
			passwordFormKey:  {helper.exampleLoginInput.Password},
			totpTokenFormKey: {helper.exampleLoginInput.TOTPToken},
		}

		var err error
		helper.req, err = http.NewRequestWithContext(
			helper.ctx,
			http.MethodPost,
			"http://todo.verygoodsoftwarenotvirus.ru",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.req.Header.Set("Content-type", "application/x-www-form-urlencoded")

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.IsType(&types.UserLoginInput{})).Return(errors.New("blah"))
		helper.service.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.service.UserLoginInputMiddleware(ms).ServeHTTP(helper.res, helper.req)

		mock.AssertExpectationsForObjects(t, ed, ms)
	})
}

func TestAuthService_AdminMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
		helper.setContextFetcher(t)

		reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.RequestContextKey, reqCtx))

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

		helper.service.AdminMiddleware(ms).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("without request context attached", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		ms := &MockHTTPHandler{}
		helper.service.AdminMiddleware(ms).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
		require.NoError(t, err)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), types.RequestContextKey, reqCtx))

		ms := &MockHTTPHandler{}
		helper.service.AdminMiddleware(ms).ServeHTTP(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})
}
