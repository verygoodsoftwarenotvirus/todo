package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/securecookie"
)

const (
	// CookieName is the name of the cookie we attach to requests.
	CookieName         = "todocookie"
	cookieErrorLogName = "_COOKIE_CONSTRUCTION_ERROR_"

	sessionInfoKey = "session_info"
)

// DecodeCookieFromRequest takes a request object and fetches the cookie data if it is present.
func (s *Service) DecodeCookieFromRequest(ctx context.Context, req *http.Request) (ca *models.SessionInfo, err error) {
	ctx, span := tracing.StartSpan(ctx, "DecodeCookieFromRequest")
	defer span.End()

	cookie, err := req.Cookie(CookieName)
	if err != http.ErrNoCookie && cookie != nil {
		var token string
		decodeErr := s.cookieManager.Decode(CookieName, cookie.Value, &token)
		if decodeErr != nil {
			return nil, fmt.Errorf("decoding request cookie: %w", decodeErr)
		}

		var sessionErr error
		ctx, sessionErr = s.sessionManager.Load(ctx, token)
		if sessionErr != nil {
			return nil, errors.New("error loading token")
		}

		si, ok := s.sessionManager.Get(ctx, sessionInfoKey).(*models.SessionInfo)
		if !ok {
			errToReturn := errors.New("no session info attached to context")
			return nil, errToReturn
		}

		return si, nil
	}

	return nil, http.ErrNoCookie
}

// WebsocketAuthFunction is provided to Newsman to determine if a user has access to websockets.
func (s *Service) WebsocketAuthFunction(req *http.Request) bool {
	ctx, span := tracing.StartSpan(req.Context(), "WebsocketAuthFunction")
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
func (s *Service) fetchUserFromCookie(ctx context.Context, req *http.Request) (*models.User, error) {
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
	ctx, span := tracing.StartSpan(req.Context(), "LoginHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("LoginHandler called")

	loginData, ok := ctx.Value(userLoginInputMiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		logger.Error(nil, "no UserLoginInput found for /login request")
		s.encoderDecoder.EncodeErrorResponse(res, "error validating request", http.StatusUnauthorized)
		return
	}

	logger = logger.WithValue("username", loginData.Username)

	user, err := s.userDB.GetUserByUsername(ctx, loginData.Username)
	if user == nil || (err != nil && err == sql.ErrNoRows) {
		logger.WithValue("user_is_nil", user == nil).Error(err, "error fetching user")
		s.encoderDecoder.EncodeErrorResponse(res, "error validating request", http.StatusUnauthorized)
		return
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachUsernameToSpan(span, user.Username)
	logger = logger.WithValue("user_id", user.ID)

	const staticError = "error encountered, please try again later"

	loginValid, err := s.validateLogin(ctx, loginData, user)
	if err != nil {
		logger.Error(err, "error encountered validating login")
		s.encoderDecoder.EncodeErrorResponse(res, staticError, http.StatusUnauthorized)
		return
	}
	logger = logger.WithValue("login_valid", loginValid)

	if !loginValid {
		logger.Debug("login was invalid")
		s.encoderDecoder.EncodeErrorResponse(res, "login was invalid", http.StatusUnauthorized)
		return
	}

	var sessionErr error
	ctx, sessionErr = s.sessionManager.Load(ctx, "")
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

	http.SetCookie(res, cookie)
	statusResponse := &models.UserStatusResponse{
		Authenticated: true,
		IsAdmin:       user.IsAdmin,
	}
	s.encoderDecoder.EncodeResponseWithStatus(res, statusResponse, http.StatusAccepted)
}

// LogoutHandler is our logout route.
func (s *Service) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "LogoutHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "error loading token")
		s.encoderDecoder.EncodeErrorResponse(res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if err := s.sessionManager.Clear(ctx); err != nil {
		logger.Error(err, "clearing user session")
		s.encoderDecoder.EncodeErrorResponse(res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if cookie, cookieRetrievalErr := req.Cookie(CookieName); cookieRetrievalErr == nil && cookie != nil {
		if c, cookieBuildingErr := s.buildCookie("deleted", time.Time{}); cookieBuildingErr == nil && c != nil {
			c.MaxAge = -1
			http.SetCookie(res, c)
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
	ctx, span := tracing.StartSpan(req.Context(), "StatusHandler")
	defer span.End()

	var (
		usr *models.UserStatusResponse
		sc  int
	)

	if userInfo, err := s.fetchUserFromCookie(ctx, req); err != nil {
		sc = http.StatusUnauthorized
		usr = &models.UserStatusResponse{
			Authenticated: false,
			IsAdmin:       false,
		}
	} else {
		sc = http.StatusOK
		usr = &models.UserStatusResponse{
			Authenticated: true,
			IsAdmin:       userInfo.IsAdmin,
		}
	}

	s.encoderDecoder.EncodeResponseWithStatus(res, usr, sc)
}

// CycleSecretHandler rotates the cookie building secret with a new random secret.
func (s *Service) CycleSecretHandler(res http.ResponseWriter, req *http.Request) {
	_, span := tracing.StartSpan(req.Context(), "CycleSecretHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Info("cycling cookie secret!")

	s.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(64),
		[]byte(s.config.CookieSecret),
	)

	res.WriteHeader(http.StatusCreated)
}

// validateLogin takes login information and returns whether or not the login is valid.
// In the event that there's an error, this function will return false and the error.
func (s *Service) validateLogin(ctx context.Context, loginInput *models.UserLoginInput, user *models.User) (bool, error) {
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

	// if the login is otherwise valid, but the password is too weak, try to rehash it.
	if err == auth.ErrCostTooLow && loginValid {
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
	} else if err != nil && err != auth.ErrCostTooLow {
		logger.Error(err, "issue validating login")
		return false, fmt.Errorf("validating login: %w", err)
	}

	return loginValid, err
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
