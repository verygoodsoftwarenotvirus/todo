package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/securecookie"
	"go.opencensus.io/trace"
)

const (
	// CookieName is the name of the cookie we attach to requests
	CookieName         = "todocookie"
	cookieErrorLogName = "_COOKIE_CONSTRUCTION_ERROR_"
)

// attachUserIDToSpan provides a consistent way to attach a userID to a given span
func attachUserIDToSpan(span *trace.Span, userID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))
	}
}

// attachUsernameToSpan provides a consistent way to attach a username to a given span
func attachUsernameToSpan(span *trace.Span, username string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("username", username))
	}
}

// DecodeCookieFromRequest takes a request object and fetches the cookie data if it is present
func (s *Service) DecodeCookieFromRequest(ctx context.Context, req *http.Request) (ca *models.CookieAuth, err error) {
	_, span := trace.StartSpan(ctx, "DecodeCookieFromRequest")
	defer span.End()

	cookie, err := req.Cookie(CookieName)
	if err != http.ErrNoCookie && cookie != nil {
		decodeErr := s.cookieManager.Decode(CookieName, cookie.Value, &ca)
		if decodeErr != nil {
			s.logger.Error(err, "decoding request cookie")
			return nil, fmt.Errorf("decoding request cookie: %w", decodeErr)
		}

		return ca, nil
	}

	return nil, http.ErrNoCookie
}

// WebsocketAuthFunction is provided to Newsman to determine if a user has access to websockets
func (s *Service) WebsocketAuthFunction(req *http.Request) bool {
	ctx, span := trace.StartSpan(req.Context(), "WebsocketAuthFunction")
	defer span.End()

	// First we check to see if there is an OAuth2 token for a valid client attached to the request.
	// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
	oauth2Client, err := s.oauth2ClientsService.ExtractOAuth2ClientFromRequest(ctx, req)
	if err == nil && oauth2Client != nil {
		return true
	}

	// In the event there's not a valid OAuth2 token attached to the request, or there is some other OAuth2 issue,
	// we next check to see if a valid cookie is attached to the request
	cookieAuth, cookieErr := s.DecodeCookieFromRequest(ctx, req)
	if cookieErr == nil && cookieAuth != nil {
		return true
	}

	// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere
	s.logger.Error(err, "error authenticated token-authenticated request")
	return false
}

// FetchUserFromRequest takes a request object and fetches the cookie, and then the user for that cookie
func (s *Service) FetchUserFromRequest(ctx context.Context, req *http.Request) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "FetchUserFromRequest")
	defer span.End()

	ca, decodeErr := s.DecodeCookieFromRequest(ctx, req)
	if decodeErr != nil {
		return nil, fmt.Errorf("fetching cookie data from request: %w", decodeErr)
	}

	user, userFetchErr := s.userDB.GetUser(req.Context(), ca.UserID)
	if userFetchErr != nil {
		return nil, fmt.Errorf("fetching user from request: %w", userFetchErr)
	}
	attachUserIDToSpan(span, ca.UserID)
	attachUsernameToSpan(span, ca.Username)

	return user, nil
}

// LoginHandler is our login route
func (s *Service) LoginHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "LoginHandler")
		defer span.End()

		loginData, errRes := s.fetchLoginDataFromRequest(req)
		if errRes != nil {
			s.logger.Error(errRes, "error encountered fetching login data from request")
			res.WriteHeader(http.StatusUnauthorized)
			if err := s.encoderDecoder.EncodeResponse(res, errRes); err != nil {
				s.logger.Error(err, "encoding response")
			}
			return
		} else if loginData == nil {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		attachUserIDToSpan(span, loginData.user.ID)
		attachUsernameToSpan(span, loginData.user.Username)

		logger := s.logger.WithValue("user", loginData.user.ID)
		loginValid, err := s.validateLogin(ctx, *loginData)
		if err != nil {
			logger.Error(err, "error encountered validating login")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger = logger.WithValue("valid", loginValid)

		if !loginValid {
			logger.Debug("login was invalid")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.Debug("login was valid")
		cookie, err := s.buildAuthCookie(loginData.user)
		if err != nil {
			logger.Error(err, "error building cookie")

			res.WriteHeader(http.StatusInternalServerError)
			response := &models.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "error encountered building cookie",
			}
			if err := s.encoderDecoder.EncodeResponse(res, response); err != nil {
				s.logger.Error(err, "encoding response")
			}
			return
		}

		http.SetCookie(res, cookie)
		res.WriteHeader(http.StatusNoContent)
	}
}

// LogoutHandler is our logout route
func (s *Service) LogoutHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		_, span := trace.StartSpan(req.Context(), "LogoutHandler")
		defer span.End()

		if cookie, err := req.Cookie(CookieName); err == nil && cookie != nil {
			c := s.buildCookie("deleted")
			c.Expires = time.Time{}
			c.MaxAge = -1
			http.SetCookie(res, c)
		} else {
			s.logger.WithError(err).Debug("logout was called, no cookie was found")
		}

		res.WriteHeader(http.StatusOK)
	}
}

// CycleSecretHandler rotates the cookie building secret with a new random secret
func (s *Service) CycleSecretHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		s.logger.Info("cycling cookie secret!")
		_, span := trace.StartSpan(req.Context(), "CycleSecretHandler")
		defer span.End()

		s.cookieManager = securecookie.New(
			securecookie.GenerateRandomKey(64),
			[]byte(s.config.CookieSecret),
		)

		res.WriteHeader(http.StatusCreated)
	}
}

type loginData struct {
	loginInput *models.UserLoginInput
	user       *models.User
}

// fetchLoginDataFromRequest searches a given HTTP request for parsed login input data, and
// returns a helper struct with the relevant login information
func (s *Service) fetchLoginDataFromRequest(req *http.Request) (*loginData, *models.ErrorResponse) {
	ctx, span := trace.StartSpan(req.Context(), "fetchLoginDataFromRequest")
	defer span.End()

	loginInput, ok := ctx.Value(UserLoginInputMiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		s.logger.Debug("no UserLoginInput found for /login request")
		return nil, &models.ErrorResponse{
			Code: http.StatusUnauthorized,
		}
	}

	username := loginInput.Username
	attachUsernameToSpan(span, username)

	// you could ensure there isn't an unsatisfied password reset token
	// requested before allowing login here

	user, err := s.userDB.GetUserByUsername(ctx, username)
	if err == sql.ErrNoRows {
		s.logger.Error(err, "no matching user")
		return nil, &models.ErrorResponse{Code: http.StatusBadRequest}
	} else if err != nil {
		s.logger.Error(err, "error fetching user")
		return nil, &models.ErrorResponse{Code: http.StatusInternalServerError}
	}
	attachUserIDToSpan(span, user.ID)

	ld := &loginData{
		loginInput: loginInput,
		user:       user,
	}

	return ld, nil
}

// validateLogin takes login information and returns whether or not the login is valid.
// In the event that there's an error, this function will return false and the error.
func (s *Service) validateLogin(ctx context.Context, loginInfo loginData) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "validateLogin")
	defer span.End()

	// alias the relevant data
	user := loginInfo.user
	loginInput := loginInfo.loginInput
	logger := s.logger.WithValue("username", user.Username)

	// check for login validity
	loginValid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
		user.Salt,
	)

	// if the login is otherwise valid, but the password is too weak, try to rehash it.
	if err == auth.ErrPasswordHashTooWeak && loginValid {
		logger.Debug("hashed password was deemed to weak, updating its hash")

		// re-hash the password
		updated, hashErr := s.authenticator.HashPassword(ctx, loginInput.Password)
		if hashErr != nil {
			return false, fmt.Errorf("updating password hash: %w", hashErr)
		}

		// update stored hashed password in the database
		user.HashedPassword = updated
		if updateErr := s.userDB.UpdateUser(ctx, user); updateErr != nil {
			return false, fmt.Errorf("saving updated password hash: %w", updateErr)
		}
	} else if err != nil && err != auth.ErrPasswordHashTooWeak {
		logger.Error(err, "issue validating login")
		return false, fmt.Errorf("validating login: %w", err)
	}

	return loginValid, nil
}

// buildAuthCookie returns an authentication cookie for a given user
func (s *Service) buildAuthCookie(user *models.User) (*http.Cookie, error) {
	// NOTE: code here is duplicated into the unit tests for
	// DecodeCookieFromRequest any changes made here might need
	// to be reflected there
	encoded, err := s.cookieManager.Encode(
		CookieName,
		models.CookieAuth{
			UserID:   user.ID,
			Admin:    user.IsAdmin,
			Username: user.Username,
		},
	)

	if err != nil {
		// NOTE: these errors should be infrequent, and should cause alarm when they do occur
		s.logger.WithName(cookieErrorLogName).WithValue("user_id", user.ID).Error(err, "error encoding cookie")
		return nil, err
	}

	return s.buildCookie(encoded), nil
}

// buildCookie provides a consistent way of constructing an HTTP cookie
func (s *Service) buildCookie(value string) *http.Cookie {
	// https://www.calhoun.io/securing-cookies-in-go/
	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.config.SecureCookiesOnly,
		Domain:   s.config.CookieDomain,
		Expires:  time.Now().Add(s.config.CookieLifetime),
	}
}
