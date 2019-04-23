package auth

import (
	"context"
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/pkg/errors"
)

const (
	// cookieName is the name of the cookie we attach to requests
	cookieName = "todocookie"
)

var (
	errNoCookie = errors.New("no cookie present in request")
)

// DecodeCookieFromRequest takes a request object and fetches the cookie data if it is present
func (s *Service) DecodeCookieFromRequest(req *http.Request) (*models.CookieAuth, error) {
	var ca *models.CookieAuth

	cookie, cerr := req.Cookie(cookieName)
	if cerr == nil && cookie != nil {
		derr := s.cookieBuilder.Decode(cookieName, cookie.Value, &ca)
		if derr != nil {
			return nil, errors.Wrap(derr, "decoding request cookie")
		}

		return ca, nil
	}
	if cerr != nil {
		return nil, cerr
	}
	return nil, errNoCookie
}

// FetchUserFromRequest takes a request object and fetches the cookie, and then the user for that cookie
func (s *Service) FetchUserFromRequest(req *http.Request) (*models.User, error) {
	ca, cerr := s.DecodeCookieFromRequest(req)
	if cerr != nil {
		return nil, errors.Wrap(cerr, "fetching cookie data from request")
	}

	user, uerr := s.database.GetUser(req.Context(), ca.UserID)
	if uerr != nil {
		return nil, errors.Wrap(uerr, "fetching user from request")
	}
	return user, nil
}

// Login is our login route
func (s *Service) Login(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	loginData, errRes := s.fetchLoginDataFromRequest(req)
	if errRes != nil {
		s.logger.Error(errRes, "error encountered fetching login data from request")
		res.WriteHeader(http.StatusUnauthorized)
		if err := s.encoder.EncodeResponse(res, errRes); err != nil {
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
	cookie, err := s.buildCookie(loginData.user)
	if err != nil {
		logger.Error(err, "error building cookie")

		res.WriteHeader(http.StatusInternalServerError)
		response := &models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "error encountered building cookie",
		}
		if err := s.encoder.EncodeResponse(res, response); err != nil {
			s.logger.Error(err, "encoding response")
		}
		return
	}

	http.SetCookie(res, cookie)
}

// Logout is our logout route
func (s *Service) Logout(res http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie(cookieName); err == nil {
		s.logger.Debug("logout was called, clearing cookie")
		cookie.MaxAge = -1
		http.SetCookie(res, cookie)
	} else {
		s.logger.WithError(err).Debug("logout was called, no cookie was found")
	}

	res.WriteHeader(http.StatusOK)
}

type loginData struct {
	loginInput *models.UserLoginInput
	user       *models.User
}

func (s *Service) fetchLoginDataFromRequest(req *http.Request) (*loginData, *models.ErrorResponse) {
	ctx := req.Context()
	loginInput, ok := ctx.Value(users.MiddlewareCtxKey).(*models.UserLoginInput)
	if !ok {
		s.logger.Debug("no UserLoginInput found for /login request")
		return nil, &models.ErrorResponse{
			Code: http.StatusUnauthorized,
		}
	}
	username := loginInput.Username
	logger := s.logger.WithValue("Username", username)

	// you could ensure there isn't an unsatisfied
	// password reset token requested before allowing login here

	user, err := s.database.GetUserByUsername(ctx, username)
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

	loginValid, err := s.authenticator.ValidateLogin(
		ctx, user.HashedPassword, loginInput.Password, user.TwoFactorSecret, loginInput.TOTPToken,
	)
	if err == auth.ErrPasswordHashTooWeak && loginValid {
		s.logger.Debug("hashed password was deemed to weak, updating its hash")

		updated, e := s.authenticator.HashPassword(ctx, loginInput.Password)
		if e != nil {
			return false, e
		}

		user.HashedPassword = updated
		if err = s.database.UpdateUser(ctx, user); err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	return loginValid, nil
}

func (s *Service) buildCookie(user *models.User) (*http.Cookie, error) {
	s.logger.WithValues(map[string]interface{}{
		"user_id": user.ID,
	}).Debug("buildCookie called")

	encoded, err := s.cookieBuilder.Encode(
		cookieName, models.CookieAuth{
			UserID:   user.ID,
			Admin:    user.IsAdmin,
			Username: user.Username,
		},
	)

	if err != nil {
		s.logger.Error(err, "error encoding cookie")
		return nil, err
	}

	// https://www.calhoun.io/securing-cookies-in-go/
	return &http.Cookie{
		Name:  cookieName,
		Value: encoded,
		// Defaults to any path on your app, but you can use this
		// to limit to a specific subdirectory.
		Path: "/",
		// true means no scripts, http requests only. This has
		// nothing to do with https vs http
		HttpOnly: true,
		// https vs http
		// Secure: true, // SECUREME
		// T // Defaults to host-only, which means exact subdomain
		// O // matching. Only change this to enable subdomains if you
		// D // need to! The below code would work on any subdomain for
		// O // yoursite.com
		///////
		// Domain: s.config.Hostname,
		// Expires: time.Now().Add(s.config.MaxCookieLifetime),
	}, nil
}