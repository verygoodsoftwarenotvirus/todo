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

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CookieAuthenticationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

		md := &mocktypes.UserDataManager{}
		md.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return(exampleUser, nil)
		s.userDB = md

		audm := &mocktypes.AccountUserMembershipDataManager{}
		audm.On("GetMembershipsForUser", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = audm

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, md, audm, ms)
	})

	T.Run("with nil user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		md := &mocktypes.UserDataManager{}
		md.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return((*types.User)(nil), nil)
		s.userDB = md

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		ms := &MockHTTPHandler{}

		s.CookieAuthenticationMiddleware(ms).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, md, ms)
	})

	T.Run("without user attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		ms := &MockHTTPHandler{}
		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ms)
	})
}

func TestService_UserAttributionMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path with cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, _, _ := fakes.BuildUserTestPrerequisites()

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("happy path with PASETO", func(t *testing.T) {
		t.SkipNow()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		s.userDB = mockDB

		audm := &mocktypes.AccountUserMembershipDataManager{}
		audm.On("GetMembershipsForUser", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return(exampleAccount.ID, examplePerms, nil)
		s.accountMembershipManager = audm

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, audm, h)
	})

	T.Run("error reading user with PASETO", func(t *testing.T) {
		t.SkipNow()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = mockDB

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.UserAttributionMiddleware(&MockHTTPHandler{}).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("error fetching relationships with PASETO", func(t *testing.T) {
		t.SkipNow()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, _ := fakes.BuildUserTestPrerequisites()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), exampleUser.ID).Return(exampleUser, nil)
		s.userDB = mockDB

		audm := &mocktypes.AccountUserMembershipDataManager{}
		audm.On("GetMembershipsForUser", mock.MatchedBy(testutil.ContextMatcher), mock.Anything).Return(exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions(nil), errors.New("blah"))
		s.accountMembershipManager = audm

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.UserAttributionMiddleware(&MockHTTPHandler{}).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, audm)
	})
}

func TestService_AuthorizationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)
		req = req.WithContext(context.WithValue(ctx, types.UserIDContextKey, reqCtx))

		s.AuthorizationMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleUser.Reputation = types.BannedAccountStatus

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)
		req = req.WithContext(context.WithValue(ctx, types.UserIDContextKey, reqCtx))

		s.AuthorizationMiddleware(&MockHTTPHandler{}).ServeHTTP(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("with missing session info", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		s.AuthorizationMiddleware(&MockHTTPHandler{}).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func Test_parseLoginInputFromForm(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		exampleUser := fakes.BuildFakeUser()
		expected := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		req.Form = map[string][]string{
			usernameFormKey:  {expected.Username},
			passwordFormKey:  {expected.Password},
			totpTokenFormKey: {expected.TOTPToken},
		}

		actual := parseLoginInputFromForm(req)
		assert.NotNil(t, actual)
		assert.Equal(t, expected, actual)
	})

	T.Run("returns nil with error parsing form", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req.URL.RawQuery = "%gh&%ij"
		req.Form = nil

		actual := parseLoginInputFromForm(req)
		assert.Nil(t, actual)
	})
}

func TestService_UserLoginInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleInput))

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
		require.NoError(t, err)
		require.NotNil(t, req)

		s := buildTestService(t)
		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.Anything).Return(errors.New("blah"))
		ed.On("EncodeErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), "attached input is invalid", http.StatusBadRequest)
		s.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ed, ms)
	})

	T.Run("with error decoding request but valid value attached to form", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		form := url.Values{
			usernameFormKey:  {exampleInput.Username},
			passwordFormKey:  {exampleInput.Password},
			totpTokenFormKey: {exampleInput.TOTPToken},
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"http://todo.verygoodsoftwarenotvirus.ru",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)
		require.NotNil(t, req)

		res := httptest.NewRecorder()
		req.Header.Set("Content-type", "application/x-www-form-urlencoded")

		s := buildTestService(t)
		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.UserLoginInputMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, ed, ms)
	})
}

func TestService_AdminMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
		exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDContextKey, reqCtx))

		s := buildTestService(t)
		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("without user attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()

		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDContextKey, reqCtx))

		s := buildTestService(t)
		ms := &MockHTTPHandler{}

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})
}
