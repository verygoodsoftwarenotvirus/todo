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

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test_parseLoginInputFromForm here

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_CookieAuthenticationMiddleware() {
	t := s.T()

	md := &mocktypes.UserDataManager{}
	md.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return(s.exampleUser, nil)
	s.service.userDataManager = md

	aumdm := &mocktypes.AccountUserMembershipDataManager{}
	aumdm.On("GetMembershipsForUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = aumdm

	ms := &MockHTTPHandler{}
	ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	_, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	s.service.CookieAuthenticationMiddleware(ms).ServeHTTP(s.res, s.req)

	mock.AssertExpectationsForObjects(t, md, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_CookieAuthenticationMiddleware_WithNilUser() {
	t := s.T()

	md := &mocktypes.UserDataManager{}
	md.On("GetUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return((*types.User)(nil), nil)
	s.service.userDataManager = md

	_, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	ms := &MockHTTPHandler{}

	s.service.CookieAuthenticationMiddleware(ms).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, md, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_CookieAuthenticationMiddleware_WithoutUserAttached() {
	t := s.T()

	ms := &MockHTTPHandler{}
	s.service.CookieAuthenticationMiddleware(ms).ServeHTTP(s.res, s.req)

	mock.AssertExpectationsForObjects(t, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_UserAttributionMiddleware() {
	t := s.T()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(s.T(), err)

	mockUserDataManager := &mocktypes.UserDataManager{}
	mockUserDataManager.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return(reqCtx, nil)
	s.service.userDataManager = mockUserDataManager

	_, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	h := &MockHTTPHandler{}
	h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.service.UserAttributionMiddleware(h).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, h)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_UserAttributionMiddleware_WithPASETO() {
	t := s.T()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(s.T(), err)

	mockDB := database.BuildMockDatabase().UserDataManager
	mockDB.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return(reqCtx, nil)
	s.service.userDataManager = mockDB

	_, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	h := &MockHTTPHandler{}
	h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.service.UserAttributionMiddleware(h).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, h)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_UserAttributionMiddleware_WithErrorFetchingRequestContextForUser() {
	t := s.T()

	mockDB := database.BuildMockDatabase().UserDataManager
	mockDB.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return((*types.RequestContext)(nil), errors.New("blah"))
	s.service.userDataManager = mockDB

	_, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	mh := &testutil.MockHTTPHandler{}
	s.service.UserAttributionMiddleware(mh).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, mockDB, mh)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_AuthorizationMiddleware() {
	t := s.T()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(s.T(), err)

	mockUserDataManager := &mocktypes.UserDataManager{}
	mockUserDataManager.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return(reqCtx, nil)
	s.service.userDataManager = mockUserDataManager

	h := &MockHTTPHandler{}
	h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.req = s.req.WithContext(context.WithValue(s.ctx, types.RequestContextKey, reqCtx))

	s.service.AuthorizationMiddleware(h).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, h)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_AuthorizationMiddleware_WithBannedUser() {
	t := s.T()

	s.exampleUser.Reputation = types.BannedAccountStatus
	s.setContextFetcher()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(s.T(), err)

	mockUserDataManager := &mocktypes.UserDataManager{}
	mockUserDataManager.On("GetRequestContextForUser", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID).Return(reqCtx, nil)
	s.service.userDataManager = mockUserDataManager

	h := &MockHTTPHandler{}
	h.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.req = s.req.WithContext(context.WithValue(s.ctx, types.RequestContextKey, reqCtx))

	mh := &testutil.MockHTTPHandler{}
	s.service.AuthorizationMiddleware(mh).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusForbidden, s.res.Code)

	mock.AssertExpectationsForObjects(t, mh)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_AuthorizationMiddleware_WithMissingRequestContext() {
	t := s.T()

	s.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return nil, nil
	}

	mh := &testutil.MockHTTPHandler{}
	s.service.AuthorizationMiddleware(mh).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, mh)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_UserLoginInputMiddleware() {
	t := s.T()

	var b bytes.Buffer
	require.NoError(t, json.NewEncoder(&b).Encode(s.exampleLoginInput))

	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", &b)
	require.NoError(t, err)
	require.NotNil(t, req)

	s.req, err = http.NewRequest(http.MethodPost, "/login", &b)
	require.NoError(t, err)

	ms := &MockHTTPHandler{}
	ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.service.UserLoginInputMiddleware(ms).ServeHTTP(s.res, s.req)

	mock.AssertExpectationsForObjects(t, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_UserLoginInputMiddleware_WithErrorDecodingRequest() {
	t := s.T()

	var b bytes.Buffer
	require.NoError(t, json.NewEncoder(&b).Encode(s.exampleLoginInput))

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.IsType(&types.UserLoginInput{})).Return(errors.New("blah"))
	ed.On("EncodeErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), "attached input is invalid", http.StatusBadRequest)
	s.service.encoderDecoder = ed

	ms := &MockHTTPHandler{}
	s.service.UserLoginInputMiddleware(ms).ServeHTTP(s.res, s.req)

	mock.AssertExpectationsForObjects(t, ed, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_UserLoginInputMiddleware_WithErrorDecodingRequestButValidValueAttachedToForm() {
	t := s.T()

	form := url.Values{
		usernameFormKey:  {s.exampleLoginInput.Username},
		passwordFormKey:  {s.exampleLoginInput.Password},
		totpTokenFormKey: {s.exampleLoginInput.TOTPToken},
	}

	var err error
	s.req, err = http.NewRequestWithContext(
		s.ctx,
		http.MethodPost,
		"http://todo.verygoodsoftwarenotvirus.ru",
		strings.NewReader(form.Encode()),
	)
	require.NoError(t, err)
	require.NotNil(t, s.req)

	s.req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.IsType(&types.UserLoginInput{})).Return(errors.New("blah"))
	s.service.encoderDecoder = ed

	ms := &MockHTTPHandler{}
	ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.service.UserLoginInputMiddleware(ms).ServeHTTP(s.res, s.req)

	mock.AssertExpectationsForObjects(t, ed, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_AdminMiddleware() {
	t := s.T()

	s.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
	s.setContextFetcher()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(s.T(), err)

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), types.RequestContextKey, reqCtx))

	ms := &MockHTTPHandler{}
	ms.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	s.service.AdminMiddleware(ms).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_AdminMiddleware_WithoutRequestContextAttached() {
	t := s.T()

	ms := &MockHTTPHandler{}

	s.service.AdminMiddleware(ms).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, ms)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_AdminMiddleware_WithNonAdminUser() {
	t := s.T()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(t, err)

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), types.RequestContextKey, reqCtx))

	ms := &MockHTTPHandler{}

	s.service.AdminMiddleware(ms).ServeHTTP(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, ms)
}
