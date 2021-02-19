package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/bcrypt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/gorilla/securecookie"
	"github.com/o1egl/paseto"
)

var (
	errNoSessionInfo = errors.New("no session info attached to context")
	errTokenLoading  = errors.New("error loading token")
)

// DecodeCookieFromRequest takes a request object and fetches the cookie data if it is present.
func (s *service) DecodeCookieFromRequest(ctx context.Context, req *http.Request) (ca *types.RequestContext, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	cookie, err := req.Cookie(s.config.Cookies.Name)
	if !errors.Is(err, http.ErrNoCookie) && cookie != nil {
		var token string

		decodeErr := s.cookieManager.Decode(s.config.Cookies.Name, cookie.Value, &token)
		if decodeErr != nil {
			return nil, fmt.Errorf("decoding request cookie: %w", decodeErr)
		}

		if ctx, err = s.sessionManager.Load(ctx, token); err != nil {
			return nil, errTokenLoading
		}

		si, ok := s.sessionManager.Get(ctx, sessionInfoKey).(*types.RequestContext)
		if !ok {
			return nil, errNoSessionInfo
		}

		return si, nil
	}

	return nil, http.ErrNoCookie
}

// fetchUserFromCookie takes a request object and fetches the cookie, and then the user for that cookie.
func (s *service) fetchUserFromCookie(ctx context.Context, req *http.Request) (*types.User, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req).WithValue("cookie_count", len(req.Cookies()))
	logger.Debug("fetchUserFromCookie called")

	ca, decodeErr := s.DecodeCookieFromRequest(ctx, req)
	if decodeErr != nil {
		s.logger.WithError(decodeErr).Debug("unable to fetch cookie data from request")
		return nil, fmt.Errorf("fetching cookie data from request: %w", decodeErr)
	}

	user, userFetchErr := s.userDB.GetUser(req.Context(), ca.User.ID)
	if userFetchErr != nil {
		s.logger.Debug("unable to determine user from request")
		return nil, fmt.Errorf("determining user from request: %w", userFetchErr)
	}

	tracing.AttachUserIDToSpan(span, ca.User.ID)

	return user, nil
}

// LoginHandler is our login route.
func (s *service) LoginHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("LoginHandler called")

	loginData, ok := ctx.Value(userLoginInputMiddlewareCtxKey).(*types.UserLoginInput)
	if !ok || loginData == nil {
		logger.Error(nil, "no UserLoginInput found for /login request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error validating request", http.StatusUnauthorized)
		return
	}

	logger = logger.WithValue(keys.UsernameKey, loginData.Username)

	user, err := s.userDB.GetUserByUsername(ctx, loginData.Username)
	if user == nil || (err != nil && errors.Is(err, sql.ErrNoRows)) {
		logger.WithValue("user_is_nil", user == nil).Error(err, "error fetching user")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error validating request", http.StatusUnauthorized)
		return
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachUsernameToSpan(span, user.Username)

	if user.IsBanned() {
		s.auditLog.LogBannedUserLoginAttemptEvent(ctx, user.ID)
		s.encoderDecoder.EncodeErrorResponse(ctx, res, user.AccountStatusExplanation, http.StatusForbidden)
		return
	}

	loginValid, err := s.validateLogin(ctx, user, loginData)
	logger = logger.WithValue(keys.UserIDKey, user.ID).WithValue("login_valid", loginValid)

	if err != nil {
		if errors.Is(err, authentication.ErrInvalidTwoFactorCode) {
			s.auditLog.LogUnsuccessfulLoginBad2FATokenEvent(ctx, user.ID)
		} else if errors.Is(err, authentication.ErrPasswordDoesNotMatch) {
			s.auditLog.LogUnsuccessfulLoginBadPasswordEvent(ctx, user.ID)
		}

		logger.Error(err, "error encountered validating login")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)

		return
	} else if !loginValid {
		logger.Debug("login was invalid")
		s.auditLog.LogUnsuccessfulLoginBadPasswordEvent(ctx, user.ID)
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "login was invalid", http.StatusUnauthorized)
		return
	}

	defaultAccount, memberships, membershipsCheckErr := s.accountMembershipManager.GetMembershipsForUser(ctx, user.ID)
	if membershipsCheckErr != nil {
		logger.Error(membershipsCheckErr, "error encountered checking for user memberships")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "error loading token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	if renewTokenErr := s.sessionManager.RenewToken(ctx); renewTokenErr != nil {
		logger.Error(err, "error encountered renewing token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	s.sessionManager.Put(ctx, sessionInfoKey, buildSessionInfoForUserLogin(user, defaultAccount, memberships))

	token, expiry, err := s.sessionManager.Commit(ctx)
	if err != nil {
		logger.Error(err, "error encountered writing to session store")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	cookie, err := s.buildCookie(token, expiry)
	if err != nil {
		logger.Error(err, "error encountered building cookie")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	logger.Debug("login successful")
	s.auditLog.LogSuccessfulLoginEvent(ctx, user.ID)
	http.SetCookie(res, cookie)

	statusResponse := user.ToStatusResponse()
	statusResponse.UserIsAuthenticated = true

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, statusResponse, http.StatusAccepted)
}

func buildSessionInfoForUserLogin(user *types.User, defaultAccount uint64, permsMap map[uint64]bitmask.ServiceUserPermissions) *types.RequestContext {
	return &types.RequestContext{
		User: types.UserRequestContext{
			Username:                user.Username,
			ID:                      user.ID,
			ActiveAccountID:         defaultAccount,
			UserAccountStatus:       user.AccountStatus,
			AccountPermissionsMap:   permsMap,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
		},
	}
}

// LogoutHandler is our logout route.
func (s *service) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "error loading token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if sessionClearErr := s.sessionManager.Clear(ctx); sessionClearErr != nil {
		logger.Error(sessionClearErr, "clearing user session")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if cookie, cookieRetrievalErr := req.Cookie(s.config.Cookies.Name); cookieRetrievalErr == nil && cookie != nil {
		if c, cookieBuildingErr := s.buildCookie("deleted", time.Time{}); cookieBuildingErr == nil && c != nil {
			c.MaxAge = -1
			http.SetCookie(res, c)
			s.auditLog.LogLogoutEvent(ctx, si.User.ID)
		} else {
			logger.Error(cookieBuildingErr, "error encountered building cookie")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "error encountered, please try again later", http.StatusInternalServerError)
			return
		}
	} else {
		logger.WithError(cookieRetrievalErr).Debug("logout was called, but encountered error loading cookie from request")
	}
}

// StatusHandler returns the user info for the user making the request.
func (s *service) StatusHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	statusResponse := &types.UserStatusResponse{}

	if user, err := s.fetchUserFromCookie(ctx, req); err == nil {
		statusResponse = user.ToStatusResponse()
		statusResponse.UserIsAuthenticated = true
	}

	s.encoderDecoder.EncodeResponse(ctx, res, statusResponse)
}

const (
	requestTimeThreshold = 2 * time.Minute
)

// PASETOHandler returns the user info for the user making the request.
func (s *service) PASETOHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	pasetoRequest, ok := ctx.Value(pasetoCreationInputMiddlewareCtxKey).(*types.PASETOCreationInput)
	if !ok || pasetoRequest == nil {
		logger.Error(nil, "no input found for PASETO request")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.DelegatedClientIDKey, pasetoRequest.ClientID)

	reqTime := time.Unix(0, pasetoRequest.RequestTime)
	if time.Until(reqTime) > requestTimeThreshold || time.Since(reqTime) > requestTimeThreshold {
		logger.WithValue("provided_request_time", reqTime.String()).Debug("PASETO request denied because its time is out of threshold")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	sum, decodeErr := base64.RawURLEncoding.DecodeString(req.Header.Get(signatureHeaderKey))
	if decodeErr != nil || len(sum) == 0 {
		logger.WithValue("sum_length", len(sum)).Error(decodeErr, "invalid signature")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	client, clientRetrievalErr := s.delegatedClientManager.GetDelegatedClient(ctx, pasetoRequest.ClientID)
	if clientRetrievalErr != nil {
		logger.Error(clientRetrievalErr, "retrieving delegated client")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	user, userRetrievalErr := s.userDB.GetUser(ctx, client.BelongsToUser)
	if userRetrievalErr != nil {
		logger.Error(userRetrievalErr, "retrieving user")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)

	defaultAccount, permissions, membershipRetrievalErr := s.accountMembershipManager.GetMembershipsForUser(ctx, client.BelongsToUser)
	if membershipRetrievalErr != nil {
		logger.Error(membershipRetrievalErr, "retrieving permissions for delegated client")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	mac := hmac.New(sha256.New, client.ClientSecret)
	if _, macWriteErr := mac.Write(s.encoderDecoder.MustJSON(pasetoRequest)); macWriteErr != nil {
		logger.Error(membershipRetrievalErr, "writing HMAC message for comparison")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	if !hmac.Equal(sum, mac.Sum(nil)) {
		logger.Info("invalid credentials passed to PASETO creation route")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	now := time.Now().UTC()

	lifetime := time.Duration(math.Min(float64(maxPASETOLifetime), float64(s.config.PASETO.Lifetime)))

	jsonToken := paseto.JSONToken{
		Audience:   strconv.FormatUint(client.BelongsToUser, 10),
		Issuer:     s.config.PASETO.Issuer,
		Jti:        uuid.New().String(),
		Subject:    strconv.FormatUint(client.BelongsToUser, 10),
		IssuedAt:   now,
		NotBefore:  now,
		Expiration: now.Add(lifetime),
	}

	si := &types.RequestContext{
		User: types.UserRequestContext{
			Username:                user.Username,
			ID:                      user.ID,
			ActiveAccountID:         defaultAccount,
			UserAccountStatus:       user.AccountStatus,
			AccountPermissionsMap:   permissions,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
		},
	}

	jsonToken.Set(pasetoDataKey, base64.RawURLEncoding.EncodeToString(si.ToBytes()))

	// Encrypt data
	token, tokenEncryptErr := paseto.NewV2().Encrypt(s.config.PASETO.LocalModeKey, jsonToken, "")
	if tokenEncryptErr != nil {
		logger.Error(tokenEncryptErr, "encrypting PASETO")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tokenRes := &types.PASETOResponse{
		Token:     token,
		ExpiresAt: jsonToken.Expiration.String(),
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, tokenRes, http.StatusAccepted)
}

// CycleCookieSecretHandler rotates the cookie building secret with a new random secret.
func (s *service) CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Info("cycling cookie secret!")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	if !si.User.ServiceAdminPermissions.CanCycleCookieSecrets() {
		logger.WithValue("admin_permissions", si.User.ServiceAdminPermissions).Debug("invalid permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	s.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(cookieSecretSize),
		[]byte(s.config.Cookies.SigningKey),
	)

	s.auditLog.LogCycleCookieSecretEvent(ctx, si.User.ID)

	res.WriteHeader(http.StatusAccepted)
}

// validateLogin takes login information and returns whether or not the login is valid.
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
		user.Salt,
	)

	if errors.Is(err, bcrypt.ErrCostTooLow) || errors.Is(err, authentication.ErrPasswordHashTooWeak) {
		// if the login is otherwise valid, but the password is too weak, try to rehash it.
		logger.Debug("hashed password was deemed to weak, updating its hash")

		// re-hash the authentication
		updated, hashErr := s.authenticator.HashPassword(ctx, loginInput.Password)
		if hashErr != nil {
			return false, fmt.Errorf("updating password hash: %w", hashErr)
		}

		// update stored hashed password in the database.
		user.HashedPassword = updated
		if updateErr := s.userDB.UpdateUser(ctx, user, nil); updateErr != nil {
			return false, fmt.Errorf("saving updated password hash: %w", updateErr)
		}

		return loginValid, nil
	}

	if errors.Is(err, authentication.ErrInvalidTwoFactorCode) || errors.Is(err, authentication.ErrPasswordDoesNotMatch) {
		return false, err
	}

	if err != nil {
		logger.Error(err, "issue validating login")
		return false, err
	}

	return loginValid, nil
}

// buildCookie provides a consistent way of constructing an HTTP cookie.
func (s *service) buildCookie(value string, expiry time.Time) (*http.Cookie, error) {
	encoded, err := s.cookieManager.Encode(s.config.Cookies.Name, value)
	if err != nil {
		// NOTE: these errors should be infrequent, and should cause alarm when they do occur
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
