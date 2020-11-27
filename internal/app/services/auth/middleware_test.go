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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

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
		exampleUser := fakes.BuildFakeUser()

		md := &mockmodels.UserDataManager{}
		md.On("GetUser", mock.Anything, mock.Anything).Return(exampleUser, nil)
		s.userDB = md

		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, md, ms)
	})

	T.Run("with nil user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)
		exampleUser := fakes.BuildFakeUser()

		md := &mockmodels.UserDataManager{}
		md.On("GetUser", mock.Anything, mock.Anything).Return((*types.User)(nil), nil)
		s.userDB = md

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		res := httptest.NewRecorder()

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		ms := &MockHTTPHandler{}
		h := s.CookieAuthenticationMiddleware(ms)
		h.ServeHTTP(res, req)

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

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return(exampleUser, nil)
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = req.WithContext(context.WithValue(ctx, types.SessionInfoKey, exampleUser.ToSessionInfo()))

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})

	T.Run("with cookie", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, mockDB, h)
	})

	T.Run("with cookie and error fetching user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		_, req = attachCookieToRequestForTest(t, s, req, exampleUser)

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		mock.AssertExpectationsForObjects(t, mockDB, h)
	})

	T.Run("with cookie present", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return(exampleUser, nil)
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = req.WithContext(context.WithValue(ctx, types.SessionInfoKey, exampleUser.ToSessionInfo()))

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})

	T.Run("with error fetching user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		ocv := &mockOAuth2ClientValidator{}
		ocv.On("ExtractOAuth2ClientFromRequest", mock.Anything, mock.Anything).Return(exampleOAuth2Client, nil)
		s.oauth2ClientsService = ocv

		mockDB := database.BuildMockDatabase().UserDataManager
		mockDB.On("GetUser", mock.Anything, exampleOAuth2Client.BelongsToUser).Return((*types.User)(nil), errors.New("blah"))
		s.userDB = mockDB

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = req.WithContext(context.WithValue(ctx, types.SessionInfoKey, exampleUser.ToSessionInfo()))

		s.UserAttributionMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ocv, mockDB, h)
	})
}

func TestService_AuthorizationMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()

		h := &MockHTTPHandler{}
		h.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = req.WithContext(context.WithValue(ctx, types.SessionInfoKey, exampleUser.ToSessionInfo()))

		s.AuthorizationMiddleware(h).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, h)
	})

	T.Run("with nil user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = req.WithContext(context.WithValue(ctx, types.SessionInfoKey, &types.SessionInfo{}))

		s.AuthorizationMiddleware(&MockHTTPHandler{}).ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleUser.AccountStatus = types.BannedAccountStatus

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		req = req.WithContext(context.WithValue(ctx, types.SessionInfoKey, exampleUser.ToSessionInfo()))

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
		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On("EncodeErrorResponse", mock.Anything, "attached input is invalid", http.StatusBadRequest)
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
		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
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

		exampleUser := fakes.BuildFakeUser()
		exampleUser.IsAdmin = true

		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				types.SessionInfoKey,
				exampleUser.ToSessionInfo(),
			),
		)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}
		ms.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

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

		exampleUser := fakes.BuildFakeUser()
		exampleUser.IsAdmin = false

		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(
				req.Context(),
				types.SessionInfoKey,
				exampleUser.ToSessionInfo(),
			),
		)

		s := buildTestService(t)
		ms := &MockHTTPHandler{}

		h := s.AdminMiddleware(ms)
		h.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)

		mock.AssertExpectationsForObjects(t, ms)
	})
}
