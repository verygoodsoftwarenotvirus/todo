package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/o1egl/paseto"
)

const (
	// pasetoCreationInputMiddlewareCtxKey is the context key for login input.
	pasetoCreationInputMiddlewareCtxKey types.ContextKey = "paseto_creation_input"
	// changeActiveAccountMiddlewareCtxKey is the context key for login input.
	changeActiveAccountMiddlewareCtxKey types.ContextKey = "change_active_account"

	signatureHeaderKey           = "Signature"
	pasetoAuthorizationHeaderKey = "Authorization"
)

var (
	errTokenExpired  = errors.New("token expired")
	errTokenNotFound = errors.New("no token data found")
)

func (s *service) fetchSessionContextDataFromPASETO(ctx context.Context, req *http.Request) (*types.SessionContextData, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	if rawToken := req.Header.Get(pasetoAuthorizationHeaderKey); rawToken != "" {
		var token paseto.JSONToken

		if err := paseto.NewV2().Decrypt(rawToken, s.config.PASETO.LocalModeKey, &token, nil); err != nil {
			return nil, observability.PrepareError(err, logger, span, "decrypting PASETO")
		}

		if time.Now().UTC().After(token.Expiration) {
			return nil, errTokenExpired
		}

		base64Encoded := token.Get(pasetoDataKey)

		gobEncoded, err := base64.RawURLEncoding.DecodeString(base64Encoded)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "decoding base64 encoded GOB payload")
		}

		var reqContext *types.SessionContextData

		if err = gob.NewDecoder(bytes.NewReader(gobEncoded)).Decode(&reqContext); err != nil {
			return nil, observability.PrepareError(err, logger, span, "decoding GOB encoded session info payload")
		}

		logger.WithValue("active_account_id", reqContext.ActiveAccountID).Debug("returning session context data")

		return reqContext, nil
	}

	return nil, errTokenNotFound
}

// CookieRequirementMiddleware requires every request have a valid cookie.
func (s *service) CookieRequirementMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		if cookie, cookieErr := req.Cookie(s.config.Cookies.Name); !errors.Is(cookieErr, http.ErrNoCookie) && cookie != nil {
			var token string
			if err := s.cookieManager.Decode(s.config.Cookies.Name, cookie.Value, &token); err == nil {
				next.ServeHTTP(res, req)
			}
		}

		// if no error was attached to the request, tell them to login first.
		http.Redirect(res, req, "/users/login", http.StatusUnauthorized)
	})
}

// UserAttributionMiddleware is concerned with figuring out who a user is, but not worried about kicking out users who are not known.
func (s *service) UserAttributionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// handle cookies if relevant.
		if cookieContext, userID, err := s.getUserIDFromCookie(ctx, req); err == nil && userID != 0 {
			logger.Debug("cookie attached to request")
			ctx = cookieContext

			tracing.AttachRequestingUserIDToSpan(span, userID)
			logger = logger.WithValue(keys.RequesterIDKey, userID)

			sessionCtxData, sessionCtxDataErr := s.accountMembershipManager.BuildSessionContextDataForUser(ctx, userID)
			if sessionCtxDataErr != nil {
				observability.AcknowledgeError(sessionCtxDataErr, logger, span, "fetching user info for cookie")
				s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
				return
			}

			s.overrideSessionContextDataValuesWithSessionData(ctx, sessionCtxData)

			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.SessionContextDataKey, sessionCtxData)))
			return
		}
		logger.Debug("no cookie attached to request")

		tokenSessionContextData, err := s.fetchSessionContextDataFromPASETO(ctx, req)
		if err != nil && !(errors.Is(err, errTokenNotFound) || errors.Is(err, errTokenExpired)) {
			observability.AcknowledgeError(err, logger, span, "extracting token from request")
		}

		if tokenSessionContextData != nil {
			// no need to fetch info since tokens are so short-lived.
			next.ServeHTTP(res, req.WithContext(context.WithValue(ctx, types.SessionContextDataKey, tokenSessionContextData)))
			return
		}

		next.ServeHTTP(res, req)
	})
}

// AuthorizationMiddleware checks to see if a user is associated with the request, and then determines whether said request can proceed.
func (s *service) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// UserAttributionMiddleware should be called before this middleware.
		if sessionCtxData, err := s.sessionContextDataFetcher(req); err == nil && sessionCtxData != nil {
			logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

			// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere.
			if sessionCtxData.Requester.Reputation == types.BannedUserReputation {
				logger.Debug("banned user attempted to make request")
				http.Redirect(res, req, "/", http.StatusForbidden)
				return
			}

			var authorizedAccounts []uint64
			for k := range sessionCtxData.AccountPermissionsMap {
				authorizedAccounts = append(authorizedAccounts, k)
			}

			logger = logger.WithValue("requested_account_id", sessionCtxData.ActiveAccountID).
				WithValue("authorized_accounts", authorizedAccounts)

			if _, authorizedForAccount := sessionCtxData.AccountPermissionsMap[sessionCtxData.ActiveAccountID]; !authorizedForAccount {
				logger.Debug("user trying to access account they are not authorized for")
				http.Redirect(res, req, "/", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(res, req)
			return
		}

		logger.Debug("no user attached to request")
		http.Redirect(res, req, "/users/login", http.StatusUnauthorized)
	})
}

// PermissionRestrictionMiddleware is concerned with figuring out who a user is, but not worried about kicking out users who are not known.
func (s *service) PermissionRestrictionMiddleware(perms ...permissions.ServiceUserPermission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx, span := s.tracer.StartSpan(req.Context())
			defer span.End()

			logger := s.logger.WithRequest(req)

			// check for a session context data first.
			sessionContextData, err := s.sessionContextDataFetcher(req)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "retrieving session context data")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			if sessionContextData.Requester.ServiceAdminPermission != 0 {
				logger.Debug("allowing admin user!")
				next.ServeHTTP(res, req)
				return
			}

			accountMemberships, allowed := sessionContextData.AccountPermissionsMap[sessionContextData.ActiveAccountID]
			if !allowed {
				logger.Debug("not authorized for account!")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			logger = logger.WithValue(keys.RequesterIDKey, sessionContextData.Requester.ID).
				WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID).
				WithValue(keys.PermissionsKey, sessionContextData.AccountPermissionsMap)

			for _, p := range perms {
				if !accountMemberships.Permissions.HasPermission(p) {
					logger.WithValue("requested_permission", p).Debug("inadequate permissions")
					s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
					return
				}
			}

			next.ServeHTTP(res, req)
		})
	}
}

// AdminMiddleware restricts requests to admin users only.
func (s *service) AdminMiddleware(next http.Handler) http.Handler {
	const staticError = "admin status required"

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "retrieving session context data")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

		if !sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin() {
			logger.Debug("AdminMiddleware called by non-admin user")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(res, req)
	})
}

// ChangeActiveAccountInputMiddleware fetches user login input from requests.
func (s *service) ChangeActiveAccountInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		x := new(types.ChangeActiveAccountInput)
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, changeActiveAccountMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// UserLoginInputMiddleware fetches user login input from requests.
func (s *service) UserLoginInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		x := new(types.UserLoginInput)
		logger := s.logger.WithRequest(req)

		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx, s.config.MinimumUsernameLength, s.config.MinimumPasswordLength); err != nil {
			observability.AcknowledgeError(err, logger, span, "validating input")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, types.UserLoginInputContextKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PASETOCreationInputMiddleware fetches user login input from requests.
func (s *service) PASETOCreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		x := new(types.PASETOCreationInput)
		if err := s.encoderDecoder.DecodeRequest(ctx, req, x); err != nil {
			observability.AcknowledgeError(err, logger, span, "decoding request body")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
			return
		}

		if err := x.ValidateWithContext(ctx); err != nil {
			logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, pasetoCreationInputMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}
