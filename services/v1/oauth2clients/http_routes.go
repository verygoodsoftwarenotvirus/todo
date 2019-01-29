package oauth2clients

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// URIParamKey is used for referring to OAuth2 client IDs in router params
	URIParamKey = "oauth2ClientID"
)

// Oauth2ClientCreationInputContextMiddleware is a middleware for attaching OAuth2 client info to a request
func (s *Service) Oauth2ClientCreationInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.Oauth2ClientCreationInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Errorf("error encountered decoding request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// Create is our OAuth2 client creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("oauth2Client creation route called")
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.Oauth2ClientCreationInput)
	if !ok {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := s.database.GetUser(input.Username)
	if err != nil {
		s.logger.Errorf("error creating oauth2Client: %v", err)
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
		s.logger.Debugln("invalid credentials provided")
		res.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		s.logger.Errorf("error validating user credentials: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	x, err := s.database.CreateOAuth2Client(input)
	if err != nil {
		s.logger.Errorf("error creating oauth2Client: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.clientStore.Set(x.ClientID, x); err != nil {
		s.logger.Errorf("error creating oauth2Client: %v", err)
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
		s.logger.Debugln("oauth2Client read route called")
		oauth2ClientID := oauth2ClientIDFetcher(req)
		i, err := s.database.GetOAuth2Client(oauth2ClientID)
		if err == sql.ErrNoRows {
			s.logger.Debugf("Read called on nonexistent client %s", oauth2ClientID)
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			s.logger.Errorf("error fetching oauth2Client %q from database: %v", oauth2ClientID, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-type", "application/json")
		json.NewEncoder(res).Encode(i)
	}
}

// List is a handler that returns a list of OAuth2 clients
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	s.logger.Debugln("oauth2Client list route called")
	qf := models.ExtractQueryFilter(req)
	oauth2Clients, err := s.database.GetOAuth2Clients(qf)
	if err != nil {
		s.logger.Errorln("encountered error getting list of oauth2 clients: ", err)
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
		s.logger.Debugln("oauth2Client deletion route called")
		oauth2ClientID := clientIDFetcher(req)

		if err := s.database.DeleteOAuth2Client(oauth2ClientID); err != nil {
			s.logger.Errorln("encountered error deleting oauth2 client: ", err)
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
		s.logger.Debugln("oauth2Client update route called")
		// input, ok := req.Context().Value(MiddlewareCtxKey).(*models.Oauth2ClientUpdateInput)
		// if !ok {
		// 	res.WriteHeader(http.StatusBadRequest)
		// 	return
		// }

		oauth2ClientID := clientIDFetcher(req)
		x, err := s.database.GetOAuth2Client(oauth2ClientID)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// IMPLEMENTME:
		//x.Update()

		if err := s.database.UpdateOAuth2Client(x); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-type", "application/json")
		json.NewEncoder(res).Encode(x)
	}
}
