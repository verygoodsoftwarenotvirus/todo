package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	errNoUserIDFoundInSession = errors.New("no user ID found in session")
)

func (s *service) overrideSessionContextDataValuesWithSessionData(ctx context.Context, sessionCtxData *types.SessionContextData) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if activeAccount, ok := s.sessionManager.Get(ctx, accountIDContextKey).(uint64); ok {
		sessionCtxData.ActiveAccountID = activeAccount
	}
}

// getUserIDFromCookie takes a request object and fetches the cookie data if it is present.
func (s *service) getUserIDFromCookie(ctx context.Context, req *http.Request) (context.Context, uint64, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	if cookie, cookieErr := req.Cookie(s.config.Cookies.Name); !errors.Is(cookieErr, http.ErrNoCookie) && cookie != nil {
		var (
			token string
			err   error
		)

		if err = s.cookieManager.Decode(s.config.Cookies.Name, cookie.Value, &token); err != nil {
			logger = logger.WithValue("cookie", cookie.Value)
			return nil, 0, observability.PrepareError(err, logger, span, "retrieving session context data")
		}

		ctx, err = s.sessionManager.Load(ctx, token)
		if err != nil {
			return nil, 0, observability.PrepareError(err, logger, span, "loading session")
		}

		if userID, ok := s.sessionManager.Get(ctx, userIDContextKey).(uint64); ok {
			logger.WithValue(keys.UserIDKey, userID).Debug("determined userID from request cookie")
			return ctx, userID, nil
		}

		return nil, 0, observability.PrepareError(errNoUserIDFoundInSession, logger, span, "determining user ID from cookie")
	}

	return nil, 0, http.ErrNoCookie
}

// determineUserFromRequestCookie takes a request object and fetches the cookie, and then the user for that cookie.
func (s *service) determineUserFromRequestCookie(ctx context.Context, req *http.Request) (*types.User, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req).WithValue("cookie_count", len(req.Cookies()))

	ctx, userID, err := s.getUserIDFromCookie(ctx, req)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching cookie data from request")
	}

	user, err := s.userDataManager.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user from database")
	}

	tracing.AttachUserIDToSpan(span, userID)
	logger.Debug("user determined from request cookie")

	return user, nil
}

// validateLogin takes login information and returns whether the login is valid.
// In the event that there's an error, this function will return false and the error.
func (s *service) validateLogin(ctx context.Context, user *types.User, loginInput *types.UserLoginInput) (bool, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	// alias the relevant data.
	logger := s.logger.WithValue(keys.UsernameKey, user.Username)

	// check for login validity.
	loginValid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
	)

	if errors.Is(err, passwords.ErrInvalidTOTPToken) || errors.Is(err, passwords.ErrPasswordDoesNotMatch) {
		return false, err
	}

	if err != nil {
		return false, observability.PrepareError(err, logger, span, "validating login")
	}

	logger.Debug("login validated")

	return loginValid, nil
}

// buildCookie provides a consistent way of constructing an HTTP cookie.
func (s *service) buildCookie(value string, expiry time.Time) (*http.Cookie, error) {
	encoded, err := s.cookieManager.Encode(s.config.Cookies.Name, value)
	if err != nil {
		// NOTE: these errs should be infrequent, and should cause alarm when they do occur
		s.logger.WithName(cookieErrorLogName).Error(err, "error encoding cookie")
		return nil, err
	}

	// https://www.calhoun.io/securing-cookies-in-go/
	cookie := &http.Cookie{
		Name:     s.config.Cookies.Name,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.config.Cookies.SecureOnly,
		Domain:   s.config.Cookies.Domain,
		Expires:  expiry,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(expiry).Seconds()),
	}

	return cookie, nil
}
