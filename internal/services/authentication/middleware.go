package authentication

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/o1egl/paseto"
)

const (
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
		if cookieContext, userID, err := s.getUserIDFromCookie(ctx, req); err == nil && userID != "" {
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
			logger = sessionCtxData.AttachToLogger(logger)

			if sessionCtxData.Requester.Reputation == types.BannedUserAccountStatus || sessionCtxData.Requester.Reputation == types.TerminatedUserReputation {
				logger.Debug("banned user attempted to make request")
				http.Redirect(res, req, "/", http.StatusForbidden)
				return
			}

			if _, authorizedForAccount := sessionCtxData.AccountPermissions[sessionCtxData.ActiveAccountID]; !authorizedForAccount {
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

// PermissionFilterMiddleware filters users out of requests based on provided functions.
func (s *service) PermissionFilterMiddleware(permissions ...authorization.Permission) func(next http.Handler) http.Handler {
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

			logger = sessionContextData.AttachToLogger(logger)

			isServiceAdmin := sessionContextData.Requester.ServicePermissions.IsServiceAdmin()

			_, allowed := sessionContextData.AccountPermissions[sessionContextData.ActiveAccountID]
			if !allowed && !isServiceAdmin {
				logger.Debug("not authorized for account!")
				s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
				return
			}

			for _, perm := range permissions {
				if !sessionContextData.ServiceRolePermissionChecker().HasPermission(perm) &&
					!sessionContextData.AccountRolePermissionsChecker().HasPermission(perm) {
					logger.WithValue("deficient_permission", perm.ID()).Debug("request filtered out")
					s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
					return
				}
			}

			next.ServeHTTP(res, req)
		})
	}
}

// ServiceAdminMiddleware restricts requests to admin users only.
func (s *service) ServiceAdminMiddleware(next http.Handler) http.Handler {
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

		logger = sessionCtxData.AttachToLogger(logger)

		if !sessionCtxData.Requester.ServicePermissions.IsServiceAdmin() {
			logger.Debug("ServiceAdminMiddleware called by non-admin user")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, staticError, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(res, req)
	})
}

var (
	// ErrNoSessionContextDataAvailable indicates no SessionContextData was attached to the request.
	ErrNoSessionContextDataAvailable = errors.New("no SessionContextData attached to session context data")
)

// FetchContextFromRequest fetches a SessionContextData from a request.
func FetchContextFromRequest(req *http.Request) (*types.SessionContextData, error) {
	if sessionCtxData, ok := req.Context().Value(types.SessionContextDataKey).(*types.SessionContextData); ok && sessionCtxData != nil {
		return sessionCtxData, nil
	}

	return nil, ErrNoSessionContextDataAvailable
}
