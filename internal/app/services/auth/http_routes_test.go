package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/bcrypt"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/gorilla/securecookie"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAuthService(t *testing.T) {
	suite.Run(t, new(authServiceHTTPRoutesTestSuite))
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_DecodeCookieFromRequest() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	_, userID, err := s.service.getUserIDFromCookie(s.ctx, s.req)
	assert.NoError(t, err)
	assert.Equal(t, s.exampleUser.ID, userID)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_DecodeCookieFromRequest_WithInvalidCookie() {
	t := s.T()

	// begin building bad cookie.
	// NOTE: any code here is duplicated from service.buildAuthCookie
	// any changes made there might need to be reflected here.
	c := &http.Cookie{
		Name:     s.service.config.Cookies.Name,
		Value:    "blah blah blah this is not a real cookie",
		Path:     "/",
		HttpOnly: true,
	}
	// end building bad cookie.
	s.req.AddCookie(c)

	_, userID, err := s.service.getUserIDFromCookie(s.req.Context(), s.req)
	assert.Error(t, err)
	assert.Zero(t, userID)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_DecodeCookieFromRequest_WithoutCookie() {
	t := s.T()

	_, userID, err := s.service.getUserIDFromCookie(s.req.Context(), s.req)
	assert.Error(t, err)
	assert.Equal(t, err, http.ErrNoCookie)
	assert.Zero(t, userID)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_determineUserFromRequestCookie() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	actualUser, err := s.service.determineUserFromRequestCookie(s.ctx, s.req)
	assert.Equal(t, s.exampleUser, actualUser)
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_determineUserFromRequestCookie_WithoutCookie() {
	t := s.T()

	actualUser, err := s.service.determineUserFromRequestCookie(s.req.Context(), s.req)
	assert.Nil(t, actualUser)
	assert.Error(t, err)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_determineUserFromRequestCookie_WithErrorFetchingUser() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	expectedError := errors.New("blah")
	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return((*types.User)(nil), expectedError)
	s.service.userDataManager = udb

	actualUser, err := s.service.determineUserFromRequestCookie(s.req.Context(), s.req)
	assert.Nil(t, actualUser)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler() {
	t := s.T()

	s.ctx = context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = authr

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = membershipDB

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On(
		"LogSuccessfulLoginEvent",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	)
	s.service.auditLog = auditLog

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusAccepted, s.res.Code)
	assert.NotEmpty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, udb, authr, auditLog, membershipDB)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithErrorGettingLoginDataFromRequest() {
	t := s.T()

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithErrorRetrievingUser() {
	t := s.T()

	s.ctx = context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return((*types.User)(nil), errors.New("blah"))
	s.service.userDataManager = udb

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithBannedUser() {
	t := s.T()

	s.exampleUser.Reputation = types.BannedAccountStatus
	s.exampleUser.ReputationExplanation = "bad behavior"

	//s.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
	//	reqCtx, _ := types.RequestContextFromUser(
	//		s.exampleUser,
	//		s.exampleAccount.ID,
	//		map[uint64]permissions.ServiceUserPermissions{},
	//	)
	//
	//	return reqCtx, nil
	//}

	s.ctx = context.WithValue(context.Background(), userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogBannedUserLoginAttemptEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID)
	s.service.auditLog = auditLog

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusForbidden, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, udb, auditLog)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithInvalidLogin() {
	t := s.T()

	s.ctx = context.WithValue(s.ctx, userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(false, nil)
	s.service.authenticator = authr

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogUnsuccessfulLoginBadPasswordEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID)
	s.service.auditLog = auditLog

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, udb, authr, auditLog)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithErrorValidatingLogin() {
	t := s.T()

	s.ctx = context.WithValue(s.ctx, userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, errors.New("blah"))
	s.service.authenticator = authr

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, udb, authr)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithErrorBuildingCookie() {
	t := s.T()

	s.ctx = context.WithValue(s.ctx, userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	cb := &mockCookieEncoderDecoder{}
	cb.On(
		"Encode",

		s.service.config.Cookies.Name,
		mock.IsType("string"),
	).Return("", errors.New("blah"))
	s.service.cookieManager = cb

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = authr

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = membershipDB

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, cb, udb, authr, membershipDB)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LoginHandler_WithErrorBuildingCookieAndErrorEncodingCookieResponse() {
	t := s.T()

	s.ctx = context.WithValue(s.ctx, userLoginInputMiddlewareCtxKey, s.exampleLoginInput)
	s.req = s.req.WithContext(s.ctx)

	cb := &mockCookieEncoderDecoder{}
	cb.On(
		"Encode",
		s.service.config.Cookies.Name,
		mock.IsType("string"),
	).Return("", errors.New("blah"))
	s.service.cookieManager = cb

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUserByUsername",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.Username,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = authr

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = membershipDB

	s.service.LoginHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)
	assert.Empty(t, s.res.Header().Get("Set-Cookie"))

	mock.AssertExpectationsForObjects(t, cb, udb, authr, membershipDB)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LogoutHandler() {
	t := s.T()

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogLogoutEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID)
	s.service.auditLog = auditLog

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	s.service.LogoutHandler(s.res, s.req)

	actualCookie := s.res.Header().Get("Set-Cookie")
	assert.Contains(t, actualCookie, "Max-Age=0")

	mock.AssertExpectationsForObjects(t, auditLog)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LogoutHandler_WithoutCookie() {
	t := s.T()

	var err error
	s.ctx, err = s.service.sessionManager.Load(s.ctx, "")
	require.NoError(t, err)
	require.NoError(t, s.service.sessionManager.RenewToken(s.ctx))

	// Then make the privilege-level change.
	s.service.sessionManager.Put(s.ctx, userIDContextKey, s.exampleUser.ID)
	s.service.sessionManager.Put(s.ctx, accountIDContextKey, s.exampleAccount.ID)

	s.service.LogoutHandler(s.res, s.req)

	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_LogoutHandler_WithErrorBuildingCookie() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)
	s.service.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(0),
		[]byte(""),
	)

	s.service.LogoutHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_validateLogin() {
	t := s.T()

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, nil)
	s.service.authenticator = authr

	actual, err := s.service.validateLogin(s.ctx, s.exampleUser, s.exampleLoginInput)
	assert.True(t, actual)
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, authr)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_validateLogin_WithTooWeakPasswordHash() {
	t := s.T()

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, authentication.ErrPasswordHashTooWeak)
	s.service.authenticator = authr

	authr.On(
		"HashPassword",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleLoginInput.Password,
	).Return("blah", nil)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"UpdateUser",
		mock.MatchedBy(testutil.ContextMatcher),
		mock.IsType(&types.User{}),
	).Return(nil)
	s.service.userDataManager = udb

	actual, err := s.service.validateLogin(s.ctx, s.exampleUser, s.exampleLoginInput)
	assert.True(t, actual)
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, authr, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_validateLogin_WithTooWeakPasswordHashAndErrorValidatingThePassword() {
	t := s.T()

	expectedErr := errors.New("arbitrary")

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, authentication.ErrPasswordHashTooWeak)

	authr.On(
		"HashPassword",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleLoginInput.Password,
	).Return("", expectedErr)
	s.service.authenticator = authr

	actual, err := s.service.validateLogin(s.ctx, s.exampleUser, s.exampleLoginInput)
	assert.False(t, actual)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, authr)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_validateLogin_WithTooWeakAPasswordHashAndErrorUpdatingUser() {
	t := s.T()

	expectedErr := errors.New("arbitrary")

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(true, bcrypt.ErrCostTooLow)

	authr.On(
		"HashPassword",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleLoginInput.Password,
	).Return("blah", nil)
	s.service.authenticator = authr

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"UpdateUser",
		mock.MatchedBy(testutil.ContextMatcher),
		mock.IsType(&types.User{}),
	).Return(expectedErr)
	s.service.userDataManager = udb

	actual, err := s.service.validateLogin(s.ctx, s.exampleUser, s.exampleLoginInput)
	assert.False(t, actual)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, authr, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_validateLogin_WithErrorValidatingLogin() { // TODO: better name
	t := s.T()

	expectedErr := errors.New("arbitrary")

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(false, expectedErr)
	s.service.authenticator = authr

	actual, err := s.service.validateLogin(s.ctx, s.exampleUser, s.exampleLoginInput)
	assert.False(t, actual)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, authr)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_validateLogin_WithInvalidLogin() {
	t := s.T()

	authr := &mockauth.Authenticator{}
	authr.On(
		"ValidateLogin",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.HashedPassword,
		s.exampleLoginInput.Password,
		s.exampleUser.TwoFactorSecret,
		s.exampleLoginInput.TOTPToken,
		s.exampleUser.Salt,
	).Return(false, nil)
	s.service.authenticator = authr

	actual, err := s.service.validateLogin(s.ctx, s.exampleUser, s.exampleLoginInput)
	assert.False(t, actual)
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, authr)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_StatusHandler() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	s.service.StatusHandler(s.res, s.req)
	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_StatusHandler_WithErrorFetchingUser() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return((*types.User)(nil), errors.New("blah"))
	s.service.userDataManager = udb

	s.service.StatusHandler(s.res, s.req)
	assert.Equal(t, http.StatusOK, s.res.Code, "expected %d in status response, got %d", http.StatusOK, s.res.Code)

	mock.AssertExpectationsForObjects(t, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_CycleSecretHandler() {
	t := s.T()

	s.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
	s.setContextFetcher()

	auditLog := &mocktypes.AuditLogEntryDataManager{}
	auditLog.On("LogCycleCookieSecretEvent", mock.MatchedBy(testutil.ContextMatcher), s.exampleUser.ID)
	s.service.auditLog = auditLog

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)
	c := s.req.Cookies()[0]

	var token string
	assert.NoError(t, s.service.cookieManager.Decode(s.service.config.Cookies.Name, c.Value, &token))

	s.service.CycleCookieSecretHandler(s.res, s.req)

	assert.Equal(t, http.StatusAccepted, s.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, s.res.Code)
	assert.Error(t, s.service.cookieManager.Decode(s.service.config.Cookies.Name, c.Value, &token))

	mock.AssertExpectationsForObjects(t, auditLog)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_CycleSecretHandler_WithErrorGettingRequestContext() {
	t := s.T()

	s.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return nil, errors.New("blah")
	}

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)
	c := s.req.Cookies()[0]

	var token string
	assert.NoError(t, s.service.cookieManager.Decode(s.service.config.Cookies.Name, c.Value, &token))

	s.service.CycleCookieSecretHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, s.res.Code)
	assert.NoError(t, s.service.cookieManager.Decode(s.service.config.Cookies.Name, c.Value, &token))
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_CycleSecretHandler_WithInvalidPermissions() {
	t := s.T()

	s.ctx, s.req = attachCookieToRequestForTest(t, s.service, s.req, s.exampleUser)
	c := s.req.Cookies()[0]

	var token string
	assert.NoError(t, s.service.cookieManager.Decode(s.service.config.Cookies.Name, c.Value, &token))

	s.service.CycleCookieSecretHandler(s.res, s.req)

	assert.Equal(t, http.StatusForbidden, s.res.Code, "expected code to be %d, but was %d", http.StatusUnauthorized, s.res.Code)
	assert.NoError(t, s.service.cookieManager.Decode(s.service.config.Cookies.Name, c.Value, &token))
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_buildCookie() {
	t := s.T()

	cookie, err := s.service.buildCookie("example", time.Now().Add(s.service.config.Cookies.Lifetime))
	assert.NotNil(t, cookie)
	assert.NoError(t, err)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_buildCookie_WithInvalidCookieBuilder() {
	t := s.T()

	s.service.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(0),
		[]byte(""),
	)

	cookie, err := s.service.buildCookie("example", time.Now().Add(s.service.config.Cookies.Lifetime))
	assert.Nil(t, cookie)
	assert.Error(t, err)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
	s.service.config.PASETO.Lifetime = time.Minute

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	expectedOutput := &types.RequestContext{
		User: types.UserRequestContext{
			ID:                      s.exampleUser.ID,
			Status:                  s.exampleUser.Reputation,
			ServiceAdminPermissions: s.exampleUser.ServiceAdminPermissions,
		},
		ActiveAccountID:       s.exampleAccount.ID,
		AccountPermissionsMap: s.examplePerms,
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientManager = dcm

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = membershipDB

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusAccepted, s.res.Code)

	// validate results

	var result *types.PASETOResponse
	require.NoError(t, json.NewDecoder(s.res.Body).Decode(&result))

	assert.NotEmpty(t, result.Token)

	var targetPayload paseto.JSONToken
	require.NoError(t, paseto.NewV2().Decrypt(result.Token, s.service.config.PASETO.LocalModeKey, &targetPayload, nil))

	assert.True(t, targetPayload.Expiration.After(time.Now().UTC()))

	payload := targetPayload.Get(pasetoDataKey)

	gobEncoding, err := base64.RawURLEncoding.DecodeString(payload)
	require.NoError(t, err)

	var si *types.RequestContext
	require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&si))

	assert.Equal(t, expectedOutput, si)

	mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_DoesNotIssueTokenWithLongerLifetimeThanPackageMaximum() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret
	s.service.config.PASETO.Lifetime = 24 * time.Hour * 365 // one year

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	expectedOutput := &types.RequestContext{
		User: types.UserRequestContext{
			ID:                      s.exampleUser.ID,
			Status:                  s.exampleUser.Reputation,
			ServiceAdminPermissions: s.exampleUser.ServiceAdminPermissions,
		},
		ActiveAccountID:       s.exampleAccount.ID,
		AccountPermissionsMap: s.examplePerms,
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientManager = dcm

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = membershipDB

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusAccepted, s.res.Code)

	// validate results

	var result *types.PASETOResponse
	require.NoError(t, json.NewDecoder(s.res.Body).Decode(&result))

	assert.NotEmpty(t, result.Token)

	var targetPayload paseto.JSONToken
	require.NoError(t, paseto.NewV2().Decrypt(result.Token, s.service.config.PASETO.LocalModeKey, &targetPayload, nil))

	assert.True(t, targetPayload.Expiration.Before(time.Now().UTC().Add(maxPASETOLifetime)))

	payload := targetPayload.Get(pasetoDataKey)

	gobEncoding, err := base64.RawURLEncoding.DecodeString(payload)
	require.NoError(t, err)

	var si *types.RequestContext
	require.NoError(t, gob.NewDecoder(bytes.NewReader(gobEncoding)).Decode(&si))

	assert.Equal(t, expectedOutput, si)

	mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithMissingInput() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithInvalidRequestTime() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: 1,
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithErrorDecodingSignatureHeader() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base32.HexEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithErrorFetchingAPIClient() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return((*types.APIClient)(nil), errors.New("blah"))
	s.service.apiClientManager = dcm

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, dcm)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithErrorFetchingUser() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientManager = dcm

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return((*types.User)(nil), errors.New("blah"))
	s.service.userDataManager = udb

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, dcm, udb)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithErrorFetchingAccountMemberships() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientManager = dcm

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(uint64(0), map[uint64]permissions.ServiceUserPermissions(nil), errors.New("blah"))
	s.service.accountMembershipManager = membershipDB

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithInvalidChecksum() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = fakes.BuildFakeAPIClient().ClientSecret

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientManager = dcm

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write([]byte("lol"))
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusUnauthorized, s.res.Code)

	mock.AssertExpectationsForObjects(t, dcm)
}

func (s *authServiceHTTPRoutesTestSuite) TestAuthService_PASETOHandler_WithTokenEncryptionError() {
	t := s.T()

	s.service.config.PASETO.LocalModeKey = nil

	exampleInput := &types.PASETOCreationInput{
		ClientID:    s.exampleAPIClient.ClientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	s.ctx = context.WithValue(context.Background(), pasetoCreationInputMiddlewareCtxKey, exampleInput)
	s.req = s.req.WithContext(s.ctx)

	dcm := &mocktypes.APIClientDataManager{}
	dcm.On(
		"GetAPIClientByClientID",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleAPIClient.ClientID,
	).Return(s.exampleAPIClient, nil)
	s.service.apiClientManager = dcm

	udb := &mocktypes.UserDataManager{}
	udb.On(
		"GetUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleUser, nil)
	s.service.userDataManager = udb

	membershipDB := &mocktypes.AccountUserMembershipDataManager{}
	membershipDB.On(
		"GetMembershipsForUser",
		mock.MatchedBy(testutil.ContextMatcher),
		s.exampleUser.ID,
	).Return(s.exampleAccount.ID, s.examplePerms, nil)
	s.service.accountMembershipManager = membershipDB

	var bodyBytes bytes.Buffer
	marshalErr := json.NewEncoder(&bodyBytes).Encode(exampleInput)
	require.NoError(t, marshalErr)

	// set HMAC signature
	mac := hmac.New(sha256.New, s.exampleAPIClient.ClientSecret)
	_, macWriteErr := mac.Write(bodyBytes.Bytes())
	require.NoError(t, macWriteErr)

	sigHeader := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	s.req.Header.Set(signatureHeaderKey, sigHeader)

	s.service.PASETOHandler(s.res, s.req)

	assert.Equal(t, http.StatusInternalServerError, s.res.Code)

	mock.AssertExpectationsForObjects(t, dcm, udb, membershipDB)
}
