package oauth2clients

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const (
	// URIParamKey is used for referring to OAuth2 client IDs in router params
	URIParamKey = "oauth2ClientID"
)

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

	x, err := s.database.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.Error(err, "error creating oauth2Client in the database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.clientStore.Set(x.ClientID, x); err != nil {
		logger.Error(err, "error setting client ID in the client store")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}

// BuildReadHandler returns a handler for retrieving an OAuth2 client
func (s *Service) BuildReadHandler(oauth2ClientIDFetcher func(req *http.Request) string) http.HandlerFunc {
	if oauth2ClientIDFetcher == nil {
		panic("oauth2ClientIDFetcher may not be nil")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		serverSpan := s.tracer.StartSpan("read route", opentracing.ChildOf(spanCtx))
		defer serverSpan.Finish()

		ctx := req.Context()
		oauth2ClientID := oauth2ClientIDFetcher(req)
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

// BuildDeleteHandler returns a Handler for deleting an OAuth2 client
func (s *Service) BuildDeleteHandler(clientIDFetcher func(req *http.Request) string) http.HandlerFunc {
	if clientIDFetcher == nil {
		panic("oauth2ClientIDFetcher may not be nil")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		serverSpan := s.tracer.StartSpan("delete route", opentracing.ChildOf(spanCtx))
		defer serverSpan.Finish()

		ctx := req.Context()
		oauth2ClientID := clientIDFetcher(req)
		logger := s.logger.WithValue("oauth2_client_id", oauth2ClientID)
		logger.Debug("oauth2Client deletion route called")

		if err := s.database.DeleteOAuth2Client(ctx, oauth2ClientID); err != nil {
			s.logger.Error(err, "encountered error deleting oauth2 client")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// BuildUpdateHandler returns a handler for updating OAuth2 clients
func (s *Service) BuildUpdateHandler(clientIDFetcher func(req *http.Request) string) http.HandlerFunc {
	if clientIDFetcher == nil {
		panic("oauth2ClientIDFetcher may not be nil")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		serverSpan := s.tracer.StartSpan("update route", opentracing.ChildOf(spanCtx))
		defer serverSpan.Finish()

		ctx := req.Context()
		oauth2ClientID := clientIDFetcher(req)
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
}
