package oauthclients

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	// "strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/go-chi/chi"
	// oauth2models "gopkg.in/oauth2.v3/models"
)

const (
	URIParamKey     = "oauth2ClientID"
	scopesSeparator = `,`
)

func (s *Oauth2ClientsService) Oauth2ClientInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.Oauth2ClientInput)
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

func (s *Oauth2ClientsService) Create(res http.ResponseWriter, req *http.Request) {
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.Oauth2ClientInput)
	if !ok {
		s.logger.Errorln("valid input not attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	x, err := s.database.CreateOauth2Client(input)
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

	// token := &oauth2models.Token{
	// 	ClientID: x.ClientID,
	// 	// UserID:
	// 	// RedirectURI
	// 	Scope: strings.Join(input.Scopes, scopesSeparator),
	// }
	// if err := s.tokenStore.Create(token); err != nil {
	// 	s.logger.Errorf("error creating oauth2Client token: %v", err)
	// 	res.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}

func (s *Oauth2ClientsService) Read(res http.ResponseWriter, req *http.Request) {
	oauth2ClientID := chi.URLParam(req, URIParamKey)
	i, err := s.database.GetOauth2Client(oauth2ClientID)
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

func (s *Oauth2ClientsService) List(res http.ResponseWriter, req *http.Request) {
	qf := models.ParseQueryFilter(req)
	oauth2Clients, err := s.database.GetOauth2Clients(qf)
	if err != nil {
		s.logger.Errorln("encountered error getting list of oauth2 clients: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(oauth2Clients)
}

func (s *Oauth2ClientsService) Delete(res http.ResponseWriter, req *http.Request) {
	// TODO: define interface for extracting these values and attach it to the service
	oauth2ClientIDParam := chi.URLParam(req, URIParamKey)
	oauth2ClientID, _ := strconv.ParseUint(oauth2ClientIDParam, 10, 64)

	if err := s.database.DeleteOauth2Client(uint(oauth2ClientID)); err != nil {
		s.logger.Errorln("encountered error deleting oauth2 client: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Update is the update route for
func (s *Oauth2ClientsService) Update(res http.ResponseWriter, req *http.Request) {
	//input, ok := req.Context().Value(MiddlewareCtxKey).(*models.Oauth2ClientInput)
	//if !ok {
	//	res.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	oauth2ClientID := chi.URLParam(req, URIParamKey)
	x, err := s.database.GetOauth2Client(oauth2ClientID)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// IMPLEMENTME:
	//x.Update()

	if err := s.database.UpdateOauth2Client(x); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	json.NewEncoder(res).Encode(x)
}
