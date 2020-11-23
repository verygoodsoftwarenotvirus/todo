package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/bcrypt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/gorilla/securecookie"
)

const (
	staticError = "error encountered, please try again later"
)

var (
	errNoSessionInfo = errors.New("no session info attached to context")
	errTokenLoading  = errors.New("error loading token")
)

// DecodeCookieFromRequest takes a request object and fetches the cookie data if it is present.
func (s *Service) DecodeCookieFromRequest(ctx context.Context, req *http.Request) (ca *types.SessionInfo, err error) {
	ctx, span := tracing.StartSpan(ctx, "DecodeCookieFromRequest")
	defer span.End()

	cookie, err := req.Cookie(CookieName)
	if !errors.Is(err, http.ErrNoCookie) && cookie != nil {
		var token string

		decodeErr := s.cookieManager.Decode(CookieName, cookie.Value, &token)
		if decodeErr != nil {
			return nil, fmt.Errorf("decoding request cookie: %w", decodeErr)
		}

		if ctx, err = s.sessionManager.Load(ctx, token); err != nil {
			return nil, errTokenLoading
		}

		si, ok := s.sessionManager.Get(ctx, sessionInfoKey).(*types.SessionInfo)
		if !ok {
			return nil, errNoSessionInfo
		}

		return si, nil
	}

	return nil, http.ErrNoCookie
}

// WebsocketAuthFunction is provided to Newsman to determine if a user has access to websockets.
func (s *Service) WebsocketAuthFunction(req *http.Request) bool {
	ctx, span := tracing.StartSpan(req.Context(), "auth.service.WebsocketAuthFunction")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// First we check to see if there is an OAuth2 token for a valid client attached to the request.
	// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
	oauth2Client, err := s.oauth2ClientsService.ExtractOAuth2ClientFromRequest(ctx, req)
	if err == nil && oauth2Client != nil {
		return true
	}

	// In the event there's not a valid OAuth2 token attached to the request, or there is some other OAuth2 issue,
	// we next check to see if a valid cookie is attached to the request.
	cookieAuth, cookieErr := s.DecodeCookieFromRequest(ctx, req)
	if cookieErr == nil && cookieAuth != nil {
		return true
	}

	// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere.
	logger.Error(err, "error authenticated token-authenticated request")
	return false
}

// fetchUserFromCookie takes a request object and fetches the cookie, and then the user for that cookie.
func (s *Service) fetchUserFromCookie(ctx context.Context, req *http.Request) (*types.User, error) {
	ctx, span := tracing.StartSpan(ctx, "fetchUserFromCookie")
	defer span.End()

	logger := s.logger.WithRequest(req).WithValue("cookie_count", len(req.Cookies()))
	logger.Debug("fetchUserFromCookie called")

	ca, decodeErr := s.DecodeCookieFromRequest(ctx, req)
	if decodeErr != nil {
		s.logger.WithError(decodeErr).Debug("unable to fetch cookie data from request")
		return nil, fmt.Errorf("fetching cookie data from request: %w", decodeErr)
	}

	user, userFetchErr := s.userDB.GetUser(req.Context(), ca.UserID)
	if userFetchErr != nil {
		s.logger.Debug("unable to determine user from request")
		return nil, fmt.Errorf("determining user from request: %w", userFetchErr)
	}

	tracing.AttachUserIDToSpan(span, ca.UserID)

	return user, nil
}

// LoginHandler is our login route.
func (s *Service) LoginHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "auth.service.LoginHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("LoginHandler called")

	loginData, ok := ctx.Value(userLoginInputMiddlewareCtxKey).(*types.UserLoginInput)
	if !ok {
		logger.Error(nil, "no UserLoginInput found for /login request")
		s.encoderDecoder.EncodeErrorResponse(res, "error validating request", http.StatusUnauthorized)
		return
	}

	logger = logger.WithValue("username", loginData.Username)

	user, err := s.userDB.GetUserByUsername(ctx, loginData.Username)
	if user == nil || (err != nil && errors.Is(err, sql.ErrNoRows)) {
		logger.WithValue("user_is_nil", user == nil).Error(err, "error fetching user")
		s.encoderDecoder.EncodeErrorResponse(res, "error validating request", http.StatusUnauthorized)
		return
	}

	if user.IsBanned() {
		s.auditLog.LogBannedUserLoginAttemptEvent(ctx, user.ID)
		s.encoderDecoder.EncodeErrorResponse(res, user.AccountStatusExplanation, http.StatusForbidden)
		return
	}

	loginValid, err := s.validateLogin(ctx, user, loginData)

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachUsernameToSpan(span, user.Username)
	logger = logger.WithValue("user_id", user.ID).WithValue("login_valid", loginValid)

	if err != nil {
		if errors.Is(err, password.ErrInvalidTwoFactorCode) {
			s.auditLog.LogUnsuccessfulLoginBad2FATokenEvent(ctx, user.ID)
		} else if errors.Is(err, password.ErrPasswordDoesNotMatch) {
			s.auditLog.LogUnsuccessfulLoginBadPasswordEvent(ctx, user.ID)
		}

		logger.Error(err, "error encountered validating login")
		s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusUnauthorized)

		return
	} else if !loginValid {
		logger.Debug("login was invalid")
		s.auditLog.LogUnsuccessfulLoginBadPasswordEvent(ctx, user.ID)
		s.encoderDecoder.EncodeErrorResponse(res, "login was invalid", http.StatusUnauthorized)
		return
	}

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "error loading token")
		s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusInternalServerError)
		return
	}

	if renewTokenErr := s.sessionManager.RenewToken(ctx); renewTokenErr != nil {
		logger.Error(err, "error encountered renewing token")
		s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusInternalServerError)
		return
	}

	s.sessionManager.Put(ctx, sessionInfoKey, user.ToSessionInfo())

	token, expiry, err := s.sessionManager.Commit(ctx)
	if err != nil {
		logger.Error(err, "error encountered writing to session store")
		s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusInternalServerError)
		return
	}

	cookie, err := s.buildCookie(token, expiry)
	if err != nil {
		logger.Error(err, "error encountered building cookie")
		s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusInternalServerError)
		return
	}

	logger.Debug("login successful")
	s.auditLog.LogSuccessfulLoginEvent(ctx, user.ID)
	http.SetCookie(res, cookie)

	statusResponse := user.ToStatusResponse()
	statusResponse.UserIsAuthenticated = true

	s.encoderDecoder.EncodeResponseWithStatus(res, statusResponse, http.StatusAccepted)
}

// LogoutHandler is our logout route.
func (s *Service) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "auth.service.LogoutHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "error loading token")
		s.encoderDecoder.EncodeErrorResponse(res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if sessionClearErr := s.sessionManager.Clear(ctx); sessionClearErr != nil {
		logger.Error(sessionClearErr, "clearing user session")
		s.encoderDecoder.EncodeErrorResponse(res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if cookie, cookieRetrievalErr := req.Cookie(CookieName); cookieRetrievalErr == nil && cookie != nil {
		if c, cookieBuildingErr := s.buildCookie("deleted", time.Time{}); cookieBuildingErr == nil && c != nil {
			c.MaxAge = -1
			http.SetCookie(res, c)
			s.auditLog.LogLogoutEvent(ctx, si.UserID)
		} else {
			logger.Error(cookieBuildingErr, "error encountered building cookie")
			s.encoderDecoder.EncodeErrorResponse(res, "error encountered, please try again later", http.StatusInternalServerError)
			return
		}
	} else {
		logger.WithError(cookieRetrievalErr).Debug("logout was called, but encountered error loading cookie from request")
	}
}

// StatusHandler returns the user info for the user making the request.
func (s *Service) StatusHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "auth.service.StatusHandler")
	defer span.End()

	statusResponse := &types.UserStatusResponse{}

	if user, err := s.fetchUserFromCookie(ctx, req); err == nil {
		statusResponse = user.ToStatusResponse()
		statusResponse.UserIsAuthenticated = true
	}

	s.encoderDecoder.EncodeResponse(res, statusResponse)
}

// CycleCookieSecretHandler rotates the cookie building secret with a new random secret.
func (s *Service) CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "auth.service.CycleCookieSecretHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Info("cycling cookie secret!")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	if !si.AdminPermissions.CanCycleCookieSecrets() {
		logger.WithValue("admin_permissions", si.AdminPermissions).Debug("invalid permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(res)
		return
	}

	s.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(cookieSecretSize),
		[]byte(s.config.CookieSigningKey),
	)

	s.auditLog.LogCycleCookieSecretEvent(ctx, si.UserID)

	res.WriteHeader(http.StatusAccepted)
}

// validateLogin takes login information and returns whether or not the login is valid.
// In the event that there's an error, this function will return false and the error.
func (s *Service) validateLogin(ctx context.Context, user *types.User, loginInput *types.UserLoginInput) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "validateLogin")
	defer span.End()

	// alias the relevant data.
	logger := s.logger.WithValue("username", user.Username)

	// check for login validity.
	loginValid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
		user.Salt,
	)

	if errors.Is(err, bcrypt.ErrCostTooLow) || errors.Is(err, password.ErrPasswordHashTooWeak) {
		// if the login is otherwise valid, but the password is too weak, try to rehash it.
		logger.Debug("hashed password was deemed to weak, updating its hash")

		// re-hash the password
		updated, hashErr := s.authenticator.HashPassword(ctx, loginInput.Password)
		if hashErr != nil {
			return false, fmt.Errorf("updating password hash: %w", hashErr)
		}

		// update stored hashed password in the database.
		user.HashedPassword = updated
		if updateErr := s.userDB.UpdateUser(ctx, user); updateErr != nil {
			return false, fmt.Errorf("saving updated password hash: %w", updateErr)
		}

		return loginValid, nil
	}

	if errors.Is(err, password.ErrInvalidTwoFactorCode) || errors.Is(err, password.ErrPasswordDoesNotMatch) {
		return false, err
	}

	if err != nil {
		logger.Error(err, "issue validating login")
		return false, err
	}

	return loginValid, nil
}

// buildCookie provides a consistent way of constructing an HTTP cookie.
func (s *Service) buildCookie(value string, expiry time.Time) (*http.Cookie, error) {
	encoded, err := s.cookieManager.Encode(CookieName, value)
	if err != nil {
		// NOTE: these errors should be infrequent, and should cause alarm when they do occur
		s.logger.WithName(cookieErrorLogName).Error(err, "error encoding cookie")
		return nil, err
	}

	// https://www.calhoun.io/securing-cookies-in-go/
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.config.SecureCookiesOnly,
		Domain:   s.config.CookieDomain,
		Expires:  expiry,
	}

	return cookie, nil
}
