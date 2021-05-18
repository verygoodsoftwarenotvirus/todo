package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"

	"github.com/google/uuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/gorilla/securecookie"
	"github.com/o1egl/paseto"
)

func (s *service) issueSessionManagedCookie(ctx context.Context, accountID, requesterID uint64) (cookie *http.Cookie, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	ctx, err = s.sessionManager.Load(ctx, "")
	if err != nil {
		// this will never happen while token is empty.
		observability.AcknowledgeError(err, logger, span, "loading token")
		return nil, err
	}

	if err = s.sessionManager.RenewToken(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "renewing token")
		return nil, err
	}

	s.sessionManager.Put(ctx, accountIDContextKey, accountID)
	s.sessionManager.Put(ctx, userIDContextKey, requesterID)

	token, expiry, err := s.sessionManager.Commit(ctx)
	if err != nil {
		// this branch cannot be tested because I cannot anticipate what the values committed will be
		observability.AcknowledgeError(err, logger, span, "writing to session store")
		return nil, err
	}

	cookie, err = s.buildCookie(token, expiry)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "building cookie")
		return nil, err
	}

	return cookie, nil
}

var (
	// ErrUserNotFound indicates a user was not located.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserBanned indicates a user is banned from using the service.
	ErrUserBanned = errors.New("user is banned")
	// ErrInvalidCredentials indicates a user provided invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (s *service) AuthenticateUser(ctx context.Context, loginData *types.UserLoginInput) (*types.User, *http.Cookie, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.UsernameKey, loginData.Username)

	user, err := s.userDataManager.GetUserByUsername(ctx, loginData.Username)
	if err != nil || user == nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, observability.PrepareError(err, logger, span, "fetching user")
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)
	tracing.AttachUserToSpan(span, user)

	if user.IsBanned() {
		s.auditLog.LogBannedUserLoginAttemptEvent(ctx, user.ID)
		return user, nil, ErrUserBanned
	}

	loginValid, err := s.validateLogin(ctx, user, loginData)
	logger.WithValue("login_valid", loginValid)

	if err != nil {
		if errors.Is(err, passwords.ErrInvalidTOTPToken) {
			s.auditLog.LogUnsuccessfulLoginBad2FATokenEvent(ctx, user.ID)
			return user, nil, ErrInvalidCredentials
		} else if errors.Is(err, passwords.ErrPasswordDoesNotMatch) {
			s.auditLog.LogUnsuccessfulLoginBadPasswordEvent(ctx, user.ID)
			return user, nil, ErrInvalidCredentials
		}

		logger.Error(err, "error encountered validating login")

		return user, nil, observability.PrepareError(err, logger, span, "validating login")
	} else if !loginValid {
		logger.Debug("login was invalid")
		s.auditLog.LogUnsuccessfulLoginBadPasswordEvent(ctx, user.ID)
		return user, nil, ErrInvalidCredentials
	}

	defaultAccountID, err := s.accountMembershipManager.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		return user, nil, observability.PrepareError(err, logger, span, "fetching user memberships")
	}

	cookie, err := s.issueSessionManagedCookie(ctx, defaultAccountID, user.ID)
	if err != nil {
		return user, nil, observability.PrepareError(err, logger, span, "issuing cookie")
	}

	s.auditLog.LogSuccessfulLoginEvent(ctx, user.ID)

	return user, cookie, nil
}

// LoginHandler is our login route.
func (s *service) LoginHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	loginData := new(types.UserLoginInput)
	if err := s.encoderDecoder.DecodeRequest(ctx, req, loginData); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err := loginData.ValidateWithContext(ctx, s.config.MinimumUsernameLength, s.config.MinimumPasswordLength); err != nil {
		observability.AcknowledgeError(err, logger, span, "validating input")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	logger = logger.WithValue(keys.UsernameKey, loginData.Username)

	user, cookie, err := s.AuthenticateUser(ctx, loginData)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		case errors.Is(err, ErrUserBanned):
			s.encoderDecoder.EncodeErrorResponse(ctx, res, user.ReputationExplanation, http.StatusForbidden)
		case errors.Is(err, ErrInvalidCredentials):
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "login was invalid", http.StatusUnauthorized)
		default:
			observability.AcknowledgeError(err, logger, span, "issuing cookie")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(res, cookie)

	statusResponse := &types.UserStatusResponse{
		UserIsAuthenticated:       true,
		UserReputation:            user.Reputation,
		UserReputationExplanation: user.ReputationExplanation,
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, statusResponse, http.StatusAccepted)
	logger.Debug("user logged in")
}

// ChangeActiveAccountHandler is our login route.
func (s *service) ChangeActiveAccountHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	input := new(types.ChangeActiveAccountInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	accountID := input.AccountID
	logger = logger.WithValue("new_session_account_id", accountID)

	requesterID := sessionCtxData.Requester.ID
	logger = logger.WithValue("user_id", requesterID)

	authorizedForAccount, err := s.accountMembershipManager.UserIsMemberOfAccount(ctx, requesterID, accountID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "checking permissions")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	if !authorizedForAccount {
		logger.Debug("invalid account ID requested for activation")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	cookie, err := s.issueSessionManagedCookie(ctx, accountID, requesterID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "issuing cookie")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusInternalServerError)
		return
	}

	logger.Info("successfully changed active session account")
	http.SetCookie(res, cookie)

	res.WriteHeader(http.StatusAccepted)
}

func (s *service) LogoutUser(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request, res http.ResponseWriter) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	ctx, err := s.sessionManager.Load(ctx, "")
	if err != nil {
		// this can literally never happen in this version of scs, because the token is empty
		return observability.PrepareError(err, logger, span, "loading token")
	}

	if destroyErr := s.sessionManager.Destroy(ctx); destroyErr != nil {
		return observability.PrepareError(err, logger, span, "destroying user session")
	}

	newCookie, cookieBuildingErr := s.buildCookie("deleted", time.Time{})
	if cookieBuildingErr != nil || newCookie == nil {
		return observability.PrepareError(cookieBuildingErr, logger, span, "building cookie")
	}

	s.auditLog.LogLogoutEvent(ctx, sessionCtxData.Requester.ID)
	newCookie.MaxAge = -1
	http.SetCookie(res, newCookie)

	logger.Debug("user logged out")

	return nil
}

// LogoutHandler is our logout route.
func (s *service) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	if err = s.LogoutUser(ctx, sessionCtxData, req, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "logging out user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

// StatusHandler returns the user info for the user making the request. TODO: DELETEME
func (s *service) StatusHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	var statusResponse *types.UserStatusResponse

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	statusResponse = &types.UserStatusResponse{
		ActiveAccount:             sessionCtxData.ActiveAccountID,
		UserReputation:            sessionCtxData.Requester.Reputation,
		UserReputationExplanation: sessionCtxData.Requester.ReputationExplanation,
		UserIsAuthenticated:       true,
	}

	s.encoderDecoder.RespondWithData(ctx, res, statusResponse)
}

const (
	pasetoRequestTimeThreshold = 2 * time.Minute
)

// PASETOHandler returns the user info for the user making the request.
func (s *service) PASETOHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	input := new(types.PASETOCreationInput)
	if err := s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	requestedAccount := input.AccountID
	logger = logger.WithValue(keys.APIClientClientIDKey, input.ClientID)

	if requestedAccount != 0 {
		logger = logger.WithValue("requested_account", requestedAccount)
	}

	reqTime := time.Unix(0, input.RequestTime)
	if time.Until(reqTime) > pasetoRequestTimeThreshold || time.Since(reqTime) > pasetoRequestTimeThreshold {
		logger.WithValue("provided_request_time", reqTime.String()).Debug("PASETO request denied because its time is out of threshold")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	sum, err := base64.RawURLEncoding.DecodeString(req.Header.Get(signatureHeaderKey))
	if err != nil || len(sum) == 0 {
		logger.WithValue("sum_length", len(sum)).Error(err, "invalid signature")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	client, clientRetrievalErr := s.apiClientManager.GetAPIClientByClientID(ctx, input.ClientID)
	if clientRetrievalErr != nil {
		observability.AcknowledgeError(err, logger, span, "fetching API client")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	mac := hmac.New(sha256.New, client.ClientSecret)
	if _, macWriteErr := mac.Write(s.encoderDecoder.MustEncodeJSON(ctx, input)); macWriteErr != nil {
		// sha256.digest.Write does not ever return an error, so this branch will remain "uncovered" :(
		observability.AcknowledgeError(err, logger, span, "writing HMAC message for comparison")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	if !hmac.Equal(sum, mac.Sum(nil)) {
		logger.Info("invalid credentials passed to PASETO creation route")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	user, err := s.userDataManager.GetUser(ctx, client.BelongsToUser)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving user")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)

	sessionCtxData, err := s.accountMembershipManager.BuildSessionContextDataForUser(ctx, user.ID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving perms for API client")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	var requestedAccountID uint64

	if requestedAccount != 0 {
		if _, isMember := sessionCtxData.AccountPermissionsMap[requestedAccount]; !isMember {
			logger.Debug("invalid account ID requested for token")
			s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
			return
		}

		logger.WithValue("requested_account", requestedAccount).Debug("setting token account ID to requested account")
		requestedAccountID = requestedAccount
		sessionCtxData.ActiveAccountID = requestedAccount
	} else {
		requestedAccountID = sessionCtxData.ActiveAccountID
	}

	logger = logger.WithValue(keys.AccountIDKey, requestedAccountID)

	// Encrypt data
	tokenRes, err := s.buildPASETOResponse(ctx, sessionCtxData, client)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "encrypting PASETO")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.Info("PASETO issued")

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, tokenRes, http.StatusAccepted)
}

func (s *service) buildPASETOToken(ctx context.Context, sessionCtxData *types.SessionContextData, client *types.APIClient) paseto.JSONToken {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	now := time.Now().UTC()
	lifetime := time.Duration(math.Min(float64(maxPASETOLifetime), float64(s.config.PASETO.Lifetime)))
	expiry := now.Add(lifetime)

	jsonToken := paseto.JSONToken{
		Audience:   strconv.FormatUint(client.BelongsToUser, 10),
		Subject:    strconv.FormatUint(client.BelongsToUser, 10),
		Jti:        uuid.NewString(),
		Issuer:     s.config.PASETO.Issuer,
		IssuedAt:   now,
		NotBefore:  now,
		Expiration: expiry,
	}

	jsonToken.Set(pasetoDataKey, base64.RawURLEncoding.EncodeToString(sessionCtxData.ToBytes()))

	return jsonToken
}

func (s *service) buildPASETOResponse(ctx context.Context, sessionCtxData *types.SessionContextData, client *types.APIClient) (*types.PASETOResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	jsonToken := s.buildPASETOToken(ctx, sessionCtxData, client)

	// Encrypt data
	token, err := paseto.NewV2().Encrypt(s.config.PASETO.LocalModeKey, jsonToken, "")
	if err != nil {
		return nil, observability.PrepareError(err, s.logger, span, "encrypting PASETO")
	}

	tokenRes := &types.PASETOResponse{
		Token:     token,
		ExpiresAt: jsonToken.Expiration.String(),
	}

	return tokenRes, nil
}

// CycleCookieSecretHandler rotates the cookie building secret with a new random secret.
func (s *service) CycleCookieSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	logger.Info("cycling cookie secret!")

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	if !sessionCtxData.Requester.ServicePermissions.CanCycleCookieSecrets() {
		logger.Debug("invalid permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	s.cookieManager = securecookie.New(
		securecookie.GenerateRandomKey(cookieSecretSize),
		[]byte(s.config.Cookies.SigningKey),
	)

	s.auditLog.LogCycleCookieSecretEvent(ctx, sessionCtxData.Requester.ID)

	res.WriteHeader(http.StatusAccepted)
}
