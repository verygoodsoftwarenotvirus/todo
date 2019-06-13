package oauth2clients

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const (
	// URIParamKey is used for referring to OAuth2 client IDs in router params
	URIParamKey = "oauth2ClientID"

	oauth2ClientIDURIParamKey                   = "client_id"
	clientIDKey               models.ContextKey = "client_id"

	scopesSeparator = ","
)

// randString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	// this is so that we don't end up with `=` in IDs
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

func (s *Service) fetchUserID(req *http.Request) uint64 {
	if x, ok := req.Context().Value(models.UserIDKey).(uint64); ok {
		return x
	}
	return 0
}

// List is a handler that returns a list of OAuth2 clients
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "list_route")
	defer span.End()

	userID := s.fetchUserID(req)
	qf := models.ExtractQueryFilter(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"filter":  qf,
		"user_id": userID,
	})
	logger.Debug("oauth2Client list route called")

	oauth2Clients, err := s.database.GetOAuth2Clients(ctx, qf, userID)
	if err == sql.ErrNoRows {
		oauth2Clients = &models.OAuth2ClientList{
			Clients: []models.OAuth2Client{},
		}
	} else if err != nil {
		logger.Error(err, "encountered error getting list of oauth2 clients from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoderDecoder.EncodeResponse(res, oauth2Clients); err != nil {
		logger.Error(err, "encoding response")
	}
}

// Create is our OAuth2 client creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "create_route")
	defer span.End()

	s.logger.Debug("oauth2Client creation route called")
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

	user, err := s.database.GetUserByUsername(ctx, input.Username)
	if err != nil {
		logger.Error(err, "fetching user by username")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	input.BelongsTo = user.ID

	valid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		input.Password,
		user.TwoFactorSecret,
		input.TOTPToken,
		user.Salt,
	)

	if !valid {
		logger.Debug("invalid credentials provided")
		res.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		logger.Error(err, "validating user credentials")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	input.ClientID = randString()
	input.ClientSecret = randString()
	input.BelongsTo = s.fetchUserID(req)

	client, err := s.database.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.Error(err, "creating oauth2Client in the database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.oauth2ClientCounter.Increment(ctx)

	logger.WithValues(map[string]interface{}{
		"client_id":       client.ID,
		"belongs_to":      client.BelongsTo,
		"client_oauth_id": client.ClientID,
	}).Debug("CreateOAuth2Client route returning successfully")

	res.WriteHeader(http.StatusCreated)
	if err = s.encoderDecoder.EncodeResponse(res, client); err != nil {
		logger.Error(err, "encoding response")
	}
}

// Read is a route handler for retrieving an OAuth2 client
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "read_route")
	defer span.End()

	userID := s.fetchUserID(req)
	oauth2ClientID := s.urlClientIDExtractor(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"oauth2_client_id": oauth2ClientID,
		"user_id":          userID,
	})
	logger.Debug("oauth2Client read route called")

	x, err := s.database.GetOAuth2Client(ctx, oauth2ClientID, userID)
	if err == sql.ErrNoRows {
		logger.Debug("Read called on nonexistent client")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error fetching oauth2Client from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
		logger.Error(err, "encoding response")
	}
}

// Delete is a route handler for deleting an OAuth2 client
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "delete_route")
	defer span.End()

	userID := s.fetchUserID(req)
	oauth2ClientID := s.urlClientIDExtractor(req)
	logger := s.logger.WithValues(map[string]interface{}{
		"oauth2_client_id": oauth2ClientID,
		"user_id":          userID,
	})
	logger.Debug("oauth2Client deletion route called")

	err := s.database.DeleteOAuth2Client(ctx, oauth2ClientID, userID)
	if err == sql.ErrNoRows {
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "encountered error deleting oauth2 client")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.oauth2ClientCounter.Decrement(ctx)

	res.WriteHeader(http.StatusNoContent)
}
