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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/gorilla/securecookie"
	"github.com/o1egl/paseto"
)

var (
	errNoSessionInfo = errors.New("no session info attached to context")
	errTokenLoading  = errors.New("error loading token")
)

// getUserIDFromCookie takes a request object and fetches the cookie data if it is present.
func (s *service) getUserIDFromCookie(ctx context.Context, req *http.Request) (context.Context, uint64, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if cookie, cookieErr := req.Cookie(s.config.Cookies.Name); !errors.Is(cookieErr, http.ErrNoCookie) && cookie != nil {
		var (
			token   string
			loadErr error
		)

		if decodeErr := s.cookieManager.Decode(s.config.Cookies.Name, cookie.Value, &token); decodeErr != nil {
			return nil, 0, fmt.Errorf("decoding request cookie: %w", decodeErr)
		}

		ctx, loadErr = s.sessionManager.Load(ctx, token)
		if loadErr != nil {
			return nil, 0, errTokenLoading
		}

		reqCtx, ok := s.sessionManager.Get(ctx, userIDContextKey).(uint64)
		if !ok {
			return nil, 0, errNoSessionInfo
		}

		return ctx, reqCtx, nil
	}

	return nil, 0, http.ErrNoCookie
}

// fetchUserFromCookie takes a request object and fetches the cookie, and then the user for that cookie.
func (s *service) fetchUserFromCookie(ctx context.Context, req *http.Request) (*types.User, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req).WithValue("cookie_count", len(req.Cookies()))
	logger.Debug("fetchUserFromCookie called")

	var (
		userID    uint64
		decodeErr error
	)

	ctx, userID, decodeErr = s.getUserIDFromCookie(ctx, req)
	if decodeErr != nil {
		s.logger.WithError(decodeErr).Debug("unable to fetch cookie data from request")
		return nil, fmt.Errorf("fetching cookie data from request: %w", decodeErr)
	}

	user, userFetchErr := s.userDataManager.GetUser(ctx, userID)
	if userFetchErr != nil {
		s.logger.Debug("unable to determine user from request")
		return nil, fmt.Errorf("determining user from request: %w", userFetchErr)
	}

	tracing.AttachUserIDToSpan(span, userID)

	return user, nil
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
		if updateErr := s.userDataManager.UpdateUser(ctx, user, nil); updateErr != nil {
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

	user, err := s.userDataManager.GetUserByUsername(ctx, loginData.Username)
	if user == nil || (err != nil && errors.Is(err, sql.ErrNoRows)) {
		logger.WithValue("user_is_nil", user == nil).Error(err, "fetching user")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error validating request", http.StatusUnauthorized)
		return
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachUsernameToSpan(span, user.Username)

	if user.IsBanned() {
		s.auditLog.LogBannedUserLoginAttemptEvent(ctx, user.ID)
		s.encoderDecoder.EncodeErrorResponse(ctx, res, user.ReputationExplanation, http.StatusForbidden)
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

	defaultAccountID, _, membershipsCheckErr := s.accountMembershipManager.GetMembershipsForUser(ctx, user.ID)
	if membershipsCheckErr != nil {
		logger.Error(membershipsCheckErr, "error encountered checking for user permissionsMap")
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
		logger.Error(renewTokenErr, "error encountered renewing token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	s.sessionManager.Put(ctx, userIDContextKey, user.ID)
	s.sessionManager.Put(ctx, accountIDContextKey, defaultAccountID)

	token, expiry, err := s.sessionManager.Commit(ctx)
	if err != nil {
		logger.Error(err, "error encountered writing to session store")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	cookie, cookieErr := s.buildCookie(token, expiry)
	if cookieErr != nil {
		logger.Error(cookieErr, "error encountered building cookie")
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

// ChangeActiveAccountHandler is our login route.
func (s *service) ChangeActiveAccountHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	input, ok := ctx.Value(changeActiveAccountMiddlewareCtxKey).(*types.ChangeActiveAccountInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	logger = logger.WithValue("requested_id", input.AccountID)

	// determine user ID.
	reqCtx, reqCtxRetrievalErr := s.requestContextFetcher(req)
	if reqCtxRetrievalErr != nil {
		logger.Error(reqCtxRetrievalErr, "error fetching request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	logger = logger.WithValue("user_id", reqCtx.User.ID)

	_, perms, permissionsErr := s.accountMembershipManager.GetMembershipsForUser(ctx, reqCtx.User.ID)
	if permissionsErr != nil {
		logger.Error(permissionsErr, "checking permissions")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	if _, authorizedForAccount := perms[input.AccountID]; !authorizedForAccount {
		logger.Debug("invalid account ID requested for activation")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "loading token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	if renewTokenErr := s.sessionManager.RenewToken(ctx); renewTokenErr != nil {
		logger.Error(renewTokenErr, "error encountered renewing token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	s.sessionManager.Put(ctx, accountIDContextKey, input.AccountID)

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

	logger.Debug("successfully changed user's cookie account")
	http.SetCookie(res, cookie)

	res.WriteHeader(http.StatusAccepted)
}

// LogoutHandler is our logout route.
func (s *service) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	ctx, sessionErr := s.sessionManager.Load(ctx, "")
	if sessionErr != nil {
		logger.Error(sessionErr, "error loading token")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if destroyErr := s.sessionManager.Destroy(ctx); destroyErr != nil {
		logger.Error(destroyErr, "destroying user session")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "error encountered, please try again later", http.StatusInternalServerError)
		return
	}

	if cookie, cookieRetrievalErr := req.Cookie(s.config.Cookies.Name); cookieRetrievalErr == nil && cookie != nil {
		if c, cookieBuildingErr := s.buildCookie("deleted", time.Time{}); cookieBuildingErr == nil && c != nil {
			c.MaxAge = -1
			http.SetCookie(res, c)
			s.auditLog.LogLogoutEvent(ctx, reqCtx.User.ID)
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

	logger = logger.WithValue(keys.APIClientClientIDKey, pasetoRequest.ClientID)

	if pasetoRequest.AccountID != 0 {
		logger = logger.WithValue("requested_account", pasetoRequest.AccountID)
	}

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

	client, clientRetrievalErr := s.apiClientManager.GetAPIClientByClientID(ctx, pasetoRequest.ClientID)
	if clientRetrievalErr != nil {
		logger.Error(clientRetrievalErr, "retrieving API client")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	mac := hmac.New(sha256.New, client.ClientSecret)
	if _, macWriteErr := mac.Write(s.encoderDecoder.MustJSON(pasetoRequest)); macWriteErr != nil {
		logger.Error(macWriteErr, "writing HMAC message for comparison")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	if !hmac.Equal(sum, mac.Sum(nil)) {
		logger.Info("invalid credentials passed to PASETO creation route")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	user, userRetrievalErr := s.userDataManager.GetUser(ctx, client.BelongsToAccount)
	if userRetrievalErr != nil {
		logger.Error(userRetrievalErr, "retrieving user")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)

	var requestedAccountID uint64

	defaultAccountID, perms, membershipRetrievalErr := s.accountMembershipManager.GetMembershipsForUser(ctx, client.BelongsToAccount)
	if membershipRetrievalErr != nil {
		logger.Error(membershipRetrievalErr, "retrieving perms for API client")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	if pasetoRequest.AccountID != 0 {
		if _, isMember := perms[pasetoRequest.AccountID]; !isMember {
			logger.WithValue("perms", perms).Debug("invalid account ID requested for token")
			s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
			return
		}

		logger.WithValue("requested_account", pasetoRequest.AccountID).Debug("setting token account ID to requested account")
		requestedAccountID = pasetoRequest.AccountID
	} else {
		requestedAccountID = defaultAccountID
	}

	logger = logger.WithValue(keys.AccountIDKey, requestedAccountID).WithValue(keys.PermissionsKey, perms)

	now := time.Now().UTC()
	lifetime := time.Duration(math.Min(float64(maxPASETOLifetime), float64(s.config.PASETO.Lifetime)))
	jsonToken := paseto.JSONToken{
		Audience:   strconv.FormatUint(client.BelongsToAccount, 10),
		Subject:    strconv.FormatUint(client.BelongsToAccount, 10),
		Jti:        uuid.New().String(),
		Issuer:     s.config.PASETO.Issuer,
		IssuedAt:   now,
		NotBefore:  now,
		Expiration: now.Add(lifetime),
	}

	reqCtx := &types.RequestContext{
		User: types.UserRequestContext{
			Username:                user.Username,
			ID:                      user.ID,
			ActiveAccountID:         requestedAccountID,
			Status:                  user.Reputation,
			AccountPermissionsMap:   perms,
			ServiceAdminPermissions: user.ServiceAdminPermissions,
		},
	}

	jsonToken.Set(pasetoDataKey, base64.RawURLEncoding.EncodeToString(reqCtx.ToBytes()))

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

	logger.Info("Issuing PASETO")

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, tokenRes, http.StatusAccepted)
}

// CycleCookieSecretHandler rotates the cookie building secret with a new random secret.
func (s *service) CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Info("cycling cookie secret!")

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	if !reqCtx.User.ServiceAdminPermissions.CanCycleCookieSecrets() {
		logger.WithValue("admin_permissions", reqCtx.User.ServiceAdminPermissions).Debug("invalid permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	s.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(cookieSecretSize),
		[]byte(s.config.Cookies.SigningKey),
	)

	s.auditLog.LogCycleCookieSecretEvent(ctx, reqCtx.User.ID)

	res.WriteHeader(http.StatusAccepted)
}
