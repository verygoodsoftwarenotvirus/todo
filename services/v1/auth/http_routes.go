package auth

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	// CookieName is the name of the cookie we attach to requests
	CookieName = "todocookie"
)

// DecodeCookieFromRequest takes a request object and fetches the cookie data if it is present
func (s *Service) DecodeCookieFromRequest(req *http.Request) (*models.CookieAuth, error) {
	var ca *models.CookieAuth

	cookie, err := req.Cookie(CookieName)
	// we don't need the nil check here, but linters won't stop complaining if we don't do it
	if err != http.ErrNoCookie && cookie != nil {
		decodeErr := s.cookieManager.Decode(CookieName, cookie.Value, &ca)
		if decodeErr != nil {
			return nil, errors.Wrap(decodeErr, "decoding request cookie")
		}

		return ca, nil
	}
	return nil, err
}

// WebsocketAuthFunction is provided to Newsman to determine if a user has access to websockets
func (s *Service) WebsocketAuthFunction(req *http.Request) bool {
	// First we check to see if there is an OAuth2 token for a valid client attached to the request.
	// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
	oauth2Client, err := s.oauth2ClientsService.RequestIsAuthenticated(req)
	if err != nil || oauth2Client == nil {
		// In the event there's not a valid OAuth2 token attached to the request, or there is some other OAuth2 issue,
		// we next check to see if a valid cookie is attached to the request
		cookieAuth, cookieErr := s.DecodeCookieFromRequest(req)
		if cookieErr != nil || cookieAuth == nil {
			// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere
			s.logger.Error(err, "error authenticated token-authenticated request")
			return false
		}

		return true
	}

	// We found a valid OAuth2 client in the request
	return true
}

// FetchUserFromRequest takes a request object and fetches the cookie, and then the user for that cookie
func (s *Service) FetchUserFromRequest(req *http.Request) (*models.User, error) {
	ca, decodeErr := s.DecodeCookieFromRequest(req)
	if decodeErr != nil {
		return nil, errors.Wrap(decodeErr, "fetching cookie data from request")
	}

	user, userFetchErr := s.userDB.GetUser(req.Context(), ca.UserID)
	if userFetchErr != nil {
		return nil, errors.Wrap(userFetchErr, "fetching user from request")
	}
	return user, nil
}

// Login is our login route
func (s *Service) Login(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "Login")
	defer span.End()

	loginData, errRes := s.fetchLoginDataFromRequest(req)
	if errRes != nil {
		s.logger.Error(errRes, "error encountered fetching login data from request")
		res.WriteHeader(http.StatusUnauthorized)
		if err := s.encoderDecoder.EncodeResponse(res, errRes); err != nil {
			s.logger.Error(err, "encoding response")
		}
		return
	}
	logger := s.logger.WithValue("login_input", loginData.loginInput)

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

	logger.Debug("login was valid, returning cookie")
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

// Logout is our logout route
func (s *Service) Logout(res http.ResponseWriter, req *http.Request) {
	_, span := trace.StartSpan(req.Context(), "Logout")
	defer span.End()

	if cookie, err := req.Cookie(CookieName); err == nil && cookie != nil {
		s.logger.Debug("logout was called, clearing cookie")
		c := s.buildCookie("deleted")
		c.Expires = time.Time{}
		c.MaxAge = -1
		http.SetCookie(res, c)
	} else {
		s.logger.WithError(err).Debug("logout was called, no cookie was found")
	}

	res.WriteHeader(http.StatusOK)
}

// CycleSecret rotates the cookie building secret with a new random secret
func (s *Service) CycleSecret(res http.ResponseWriter, req *http.Request) {
	_, span := trace.StartSpan(req.Context(), "CycleSecret")
	defer span.End()

	s.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(64),
		[]byte(s.config.CookieSecret),
	)

	res.WriteHeader(http.StatusCreated)
}

type loginData struct {
	loginInput *models.UserLoginInput
	user       *models.User
}

func (s *Service) fetchLoginDataFromRequest(req *http.Request) (*loginData, *models.ErrorResponse) {
	ctx := req.Context()
	loginInput, ok := ctx.Value(UserLoginInputMiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		s.logger.Debug("no UserLoginInput found for /login request")
		return nil, &models.ErrorResponse{
			Code: http.StatusUnauthorized,
		}
	}
	username := loginInput.Username
	logger := s.logger.WithValue("Username", username)

	// you could ensure there isn't an unsatisfied password reset token
	// requested before allowing login here

	user, err := s.userDB.GetUserByUsername(ctx, username)
	if err == sql.ErrNoRows {
		logger.WithError(err).Debug("no matching user")
		return nil, &models.ErrorResponse{
			Code: http.StatusBadRequest,
		}
	} else if err != nil {
		logger.WithError(err).Debug("error fetching user")
		return nil, &models.ErrorResponse{
			Code: http.StatusInternalServerError,
		}
	}

	ld := &loginData{
		loginInput: loginInput,
		user:       user,
	}

	return ld, nil
}

func (s *Service) validateLogin(ctx context.Context, loginInfo loginData) (bool, error) {
	user := loginInfo.user
	loginInput := loginInfo.loginInput

	logger := s.logger.WithValue("2fa_secret", user.TwoFactorSecret)
	logger.Debug("validating login")

	loginValid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
		user.Salt,
	)
	if err == auth.ErrPasswordHashTooWeak && loginValid {
		s.logger.Debug("hashed password was deemed to weak, updating its hash")

		updated, e := s.authenticator.HashPassword(ctx, loginInput.Password)
		if e != nil {
			return false, e
		}

		user.HashedPassword = updated
		if updateErr := s.userDB.UpdateUser(ctx, user); updateErr != nil {
			return false, updateErr
		}
	} else if err != nil && err != auth.ErrPasswordHashTooWeak {
		logger.Error(err, "issue validating login")
		return false, err
	}

	return loginValid, nil
}

func (s *Service) buildAuthCookie(user *models.User) (*http.Cookie, error) {
	s.logger.WithValues(map[string]interface{}{
		"user_id": user.ID,
	}).Debug("buildAuthCookie called")

	// NOTE: code here is duplicated into the unit tests for DecodeCookieFromRequest
	// any changes made here might need to be reflected there
	encoded, err := s.cookieManager.Encode(
		CookieName, models.CookieAuth{
			UserID:   user.ID,
			Admin:    user.IsAdmin,
			Username: user.Username,
		},
	)

	if err != nil {
		s.logger.Error(err, "error encoding cookie")
		return nil, err
	}

	return s.buildCookie(encoded), nil
}

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
