package oauth2clients

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2server "gopkg.in/oauth2.v3/server"
)

const (
	// URIParamKey is used for referring to OAuth2 client IDs in router params
	URIParamKey = "oauth2ClientID"

	scopesKey   models.ContextKey = "scopes"
	clientIDKey models.ContextKey = "client_id"

	scopesSeparator           = ","
	oauth2ClientIDURIParamKey = "client_id"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

// randString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() (string, error) {
	b := make([]byte, 64)
	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

// OAuth2ClientCreationInputContextMiddleware is a middleware for attaching OAuth2 client info to a request
func (s *Service) OAuth2ClientCreationInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.OAuth2ClientCreationInput)

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// Create is our OAuth2 client creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("create route", opentracing.ChildOf(spanCtx))
	defer serverSpan.Finish()

	s.logger.Debug("oauth2Client creation route called")
	ctx := req.Context()
	input, ok := ctx.Value(MiddlewareCtxKey).(*models.OAuth2ClientCreationInput)
	if !ok {
		s.logger.Error(nil, "valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	logger := s.logger.WithValues(map[string]interface{}{
		"username":     input.Username,
		"scopes":       input.Scopes,
		"redirect_uri": input.RedirectURI,
	})

	user, err := s.database.GetUser(ctx, input.Username)
	if err != nil {
		logger.Error(err, "error creating oauth2Client")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	input.BelongsTo = user.ID

	if valid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		input.Password,
		user.TwoFactorSecret,
		input.TOTPToken,
	); !valid {
		logger.Debug("invalid credentials provided")
		res.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		logger.Error(err, "error validating user credentials")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if input.ClientID, err = randString(); err != nil {
		logger.Error(err, "generating OAuth2 Client ID")
		return
	} else if input.ClientSecret, err = randString(); err != nil {
		logger.Error(err, "generating OAuth2 Client Secret")
		return
	}

	x, err := s.database.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.Error(err, "error creating oauth2Client in the database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.oauth2ClientStore.Set(x.ClientID, x); err != nil {
		logger.Error(err, "error setting client ID in the client store")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}

// Read is a route handler for retrieving an OAuth2 client
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("read route", opentracing.ChildOf(spanCtx))
	defer serverSpan.Finish()

	ctx := req.Context()
	oauth2ClientID := s.clientIDFetcher(req)
	logger := s.logger.WithValue("oauth2_client_id", oauth2ClientID)
	logger.Debug("oauth2Client read route called")

	i, err := s.database.GetOAuth2Client(ctx, oauth2ClientID)
	if err == sql.ErrNoRows {
		logger.Debug("Read called on nonexistent client")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error fetching oauth2Client from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(i)
}

// List is a handler that returns a list of OAuth2 clients
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("list route", opentracing.ChildOf(spanCtx))
	defer serverSpan.Finish()

	ctx := req.Context()
	qf := models.ExtractQueryFilter(req)
	logger := s.logger.WithValue("filter", qf)
	logger.Debug("oauth2Client list route called")

	oauth2Clients, err := s.database.GetOAuth2Clients(ctx, qf)
	if err != nil {
		logger.Error(err, "encountered error getting list of oauth2 clients from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(oauth2Clients)
}

// Delete is a route handler for deleting an OAuth2 client
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("delete route", opentracing.ChildOf(spanCtx))
	defer serverSpan.Finish()

	ctx := req.Context()
	oauth2ClientID := s.clientIDFetcher(req)
	logger := s.logger.WithValue("oauth2_client_id", oauth2ClientID)
	logger.Debug("oauth2Client deletion route called")

	if err := s.database.DeleteOAuth2Client(ctx, oauth2ClientID); err != nil {
		s.logger.Error(err, "encountered error deleting oauth2 client")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusNoContent)
}

// Update is a route handler for updating OAuth2 clients
func (s *Service) Update(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("update route", opentracing.ChildOf(spanCtx))
	defer serverSpan.Finish()

	ctx := req.Context()
	oauth2ClientID := s.clientIDFetcher(req)
	logger := s.logger.WithValue("oauth2_client_id", oauth2ClientID)
	logger.Debug("oauth2Client update route called")
	// input, ok := req.Context().Value(MiddlewareCtxKey).(*models.OAuth2ClientUpdateInput)
	// if !ok {
	// 	res.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	x, err := s.database.GetOAuth2Client(ctx, oauth2ClientID)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// IMPLEMENTME:
	//x.Update()

	if err := s.database.UpdateOAuth2Client(ctx, x); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.InternalErrorHandler = (*Service)(nil).OAuth2InternalErrorHandler

// OAuth2InternalErrorHandler fulfills a role for the OAuth2 server-side provider
func (s *Service) OAuth2InternalErrorHandler(err error) *oauth2errors.Response {
	res := &oauth2errors.Response{
		Error:       err,
		Description: "Internal error",
		ErrorCode:   http.StatusInternalServerError,
		StatusCode:  http.StatusInternalServerError,
	}

	s.logger.Error(err, "OAuth2 Internal Error")
	return res
}

var _ oauth2server.ResponseErrorHandler = (*Service)(nil).OAuth2ResponseErrorHandler

// OAuth2ResponseErrorHandler fulfills a role for the OAuth2 server-side provider
func (s *Service) OAuth2ResponseErrorHandler(re *oauth2errors.Response) {
	s.logger.WithValues(map[string]interface{}{
		"error":       re.Error,
		"error_code":  re.ErrorCode,
		"description": re.Description,
		"URI":         re.URI,
		"status_code": re.StatusCode,
		"header":      re.Header,
	})
}

// OauthTokenAuthenticationMiddleware authenticates Oauth tokens
func (s *Service) OauthTokenAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("OauthTokenAuthenticationMiddleware called")

		token, err := s.oauth2Handler.ValidationBearerToken(req)
		if err != nil || token == nil {
			s.logger.Error(err, "error validating bearer token")
			http.Error(res, "invalid token", http.StatusUnauthorized)
			return
		}

		cid := token.GetClientID()
		logger := s.logger.WithValue("client_id", cid)

		c, err := s.database.GetOAuth2Client(ctx, cid)
		if err != nil {
			logger.Error(err, "error fetching OAuth2 Client")
			http.Error(res, fmt.Sprintf("error fetching client ID: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		req = req.WithContext(context.WithValue(ctx, models.UserIDKey, c.BelongsTo))
		next.ServeHTTP(res, req)
	})
}

// OAuth2ClientInfoMiddleware fetches clientOAuth2Client info from requests and attaches it eplicitly to a request
func (s *Service) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("OauthInfoMiddleware called")

		if v := req.URL.Query().Get(oauth2ClientIDURIParamKey); v != "" {
			logger := s.logger.WithValue("oauth2_client_id", v)
			logger.Debug("fetching oauth2 client from database")

			client, err := s.database.GetOAuth2Client(ctx, v)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 client")
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Service) fetchOAuth2ClientFromRequest(req *http.Request) *models.OAuth2Client {
	s.logger.Debug("fetchOAuth2ClientFromRequest called")
	ctx := req.Context()
	client, ok := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client)
	if !ok {
		return nil
	}
	return client
}

func (s *Service) fetchOAuth2ClientScopesFromRequest(req *http.Request) []string {
	s.logger.Debug("fetchOAuth2ClientScopesFromRequest called")
	ctx := req.Context()
	scopes, ok := ctx.Value(scopesKey).([]string)
	if !ok {
		return nil
	}
	return scopes
}

func (s *Service) fetchOAuth2ClientIDFromRequest(req *http.Request) string {
	s.logger.Debug("fetchOAuth2ClientIDFromRequest called")
	ctx := req.Context()
	clientID, ok := ctx.Value(clientIDKey).(string)
	if !ok {
		return ""
	}
	return clientID
}

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.AuthorizeScopeHandler = (*Service)(nil).AuthorizeScopeHandler

// AuthorizeScopeHandler satisfies the oauth2server AuthorizeScopeHandler interface
func (s *Service) AuthorizeScopeHandler(res http.ResponseWriter, req *http.Request) (scope string, err error) {
	ctx := req.Context()

	s.logger.Debug("AuthorizeScopeHandler called")
	client := s.fetchOAuth2ClientFromRequest(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"client.ID": client.ID,
	})

	if client == nil {
		clientID := s.fetchOAuth2ClientIDFromRequest(req)
		if clientID != "" {
			client, err := s.database.GetOAuth2Client(ctx, clientID)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 Client")
				return "", err
			}

			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
			return strings.Join(client.Scopes, scopesSeparator), nil
		}
	} else {
		return strings.Join(client.Scopes, scopesSeparator), nil
	}

	returnErr := errors.New("no scope information found")
	logger.Error(nil, "no scope information found")
	return "", returnErr
}

var _ oauth2server.UserAuthorizationHandler = (*Service)(nil).UserAuthorizationHandler

// UserAuthorizationHandler satisfies the oauth2server UserAuthorizationHandler interface
func (s *Service) UserAuthorizationHandler(res http.ResponseWriter, req *http.Request) (userID string, err error) {
	ctx := req.Context()
	s.logger.Debug("UserAuthorizationHandler called")

	var uid uint64
	if client, clientOk := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client); !clientOk {
		user, ok := ctx.Value(models.UserKey).(*models.User)
		if !ok {
			s.logger.Debug("no user attached to this request")
			return "", errors.New("user not found")
		}
		uid = user.ID
	} else {
		uid = client.BelongsTo
	}
	return strconv.FormatUint(uid, 10), nil
}

var _ oauth2server.ClientAuthorizedHandler = (*Service)(nil).ClientAuthorizedHandler

// ClientAuthorizedHandler satisfies the oauth2server ClientAuthorizedHandler interface
func (s *Service) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	s.logger.Debug("ClientAuthorizedHandler called")

	if grant == oauth2.PasswordCredentials {
		return false, errors.New("invalid grant type: password")
	}
	client, err := s.database.GetOAuth2Client(context.Background(), clientID)
	if err != nil {
		return false, err
	}
	// FINISHME: what if client is deactivated?!

	if grant == oauth2.Implicit && !client.ImplicitAllowed {
		return false, errors.New("client not authorized for implicit grants")
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*Service)(nil).ClientScopeHandler

// ClientScopeHandler satisfies the oauth2server ClientScopeHandler interface
func (s *Service) ClientScopeHandler(clientID, scope string) (authed bool, err error) {
	logger := s.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
		"scope":     scope,
	})
	logger.Debug("ClientScopeHandler called")

	c, err := s.database.GetOAuth2Client(context.Background(), clientID)
	if err != nil {
		logger.Error(err, "error fetching OAuth2 client for ClientScopeHandler")
		return false, err
	}

	logger = logger.WithValue("oauth2_client_scopes", c.Scopes)
	logger.Debug("OAuth2 Client retrieved in ClientScopeHandler")

	for _, s := range c.Scopes {
		if s == scope || s == "*" {
			authed = true
		}
	}

	logger.WithValue("authed", authed).Debug("returning from ClientScopeHandler")
	return authed, nil
}
