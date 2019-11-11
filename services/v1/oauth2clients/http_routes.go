package oauth2clients

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"net/http"
	"strconv"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

const (
	// URIParamKey is used for referring to OAuth2 client IDs in router params
	URIParamKey = "oauth2ClientID"

	oauth2ClientIDURIParamKey                   = "client_id"
	clientIDKey               models.ContextKey = "client_id"
)

// attachUserIDToSpan provides a consistent way of attaching an user ID to a given span
func attachUserIDToSpan(span *trace.Span, userID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))
	}
}

// attachOAuth2ClientDatabaseIDToSpan provides a consistent way of attaching an oauth2 client ID to a given span
func attachOAuth2ClientDatabaseIDToSpan(span *trace.Span, clientID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("oauth2client_db_id", strconv.FormatUint(clientID, 10)))
	}
}

// attachOAuth2ClientIDToSpan provides a consistent way of attaching a client ID to a given span
func attachOAuth2ClientIDToSpan(span *trace.Span, clientID string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("client_id", clientID))
	}
}

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

// fetchUserID grabs a userID out of the request context
func (s *Service) fetchUserID(req *http.Request) uint64 {
	if id, ok := req.Context().Value(models.UserIDKey).(uint64); ok {
		return id
	}
	return 0
}

// ListHandler is a handler that returns a list of OAuth2 clients
func (s *Service) ListHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ListHandler")
		defer span.End()

		// extract filter
		qf := models.ExtractQueryFilter(req)

		// determine user
		userID := s.fetchUserID(req)
		attachUserIDToSpan(span, userID)
		logger := s.logger.WithValue("user_id", userID)

		// fetch oauth2 clients
		oauth2Clients, err := s.database.GetOAuth2Clients(ctx, qf, userID)
		if err == sql.ErrNoRows {
			// just return an empty list if there are no results
			oauth2Clients = &models.OAuth2ClientList{
				Clients: []models.OAuth2Client{},
			}
		} else if err != nil {
			logger.Error(err, "encountered error getting list of oauth2 clients from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode response and peace
		if err = s.encoderDecoder.EncodeResponse(res, oauth2Clients); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// CreateHandler is our OAuth2 client creation route
func (s *Service) CreateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "CreateHandler")
		defer span.End()

		// fetch creation input from request context
		input, ok := ctx.Value(CreationMiddlewareCtxKey).(*models.OAuth2ClientCreationInput)
		if !ok {
			s.logger.Info("valid input not attached to request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// keep relevant data in mind
		logger := s.logger.WithValues(map[string]interface{}{
			"username":     input.Username,
			"scopes":       input.Scopes,
			"redirect_uri": input.RedirectURI,
		})

		// retrieve user
		user, err := s.database.GetUserByUsername(ctx, input.Username)
		if err != nil {
			logger.Error(err, "fetching user by username")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		input.BelongsTo = user.ID

		// tag span since we have the info
		attachUserIDToSpan(span, user.ID)

		// check credentials
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

		// set some data
		input.ClientID = randString()
		input.ClientSecret = randString()
		input.BelongsTo = s.fetchUserID(req)

		// create the client
		client, err := s.database.CreateOAuth2Client(ctx, input)
		if err != nil {
			logger.Error(err, "creating oauth2Client in the database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify interested parties
		attachOAuth2ClientDatabaseIDToSpan(span, client.ID)
		s.oauth2ClientCounter.Increment(ctx)

		res.WriteHeader(http.StatusCreated)
		if err = s.encoderDecoder.EncodeResponse(res, client); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// ReadHandler is a route handler for retrieving an OAuth2 client
func (s *Service) ReadHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ReadHandler")
		defer span.End()

		// determine subject of request
		userID := s.fetchUserID(req)
		oauth2ClientID := s.urlClientIDExtractor(req)

		// keep the aforementioned in mind
		logger := s.logger.WithValues(map[string]interface{}{
			"oauth2_client_id": oauth2ClientID,
			"user_id":          userID,
		})
		attachUserIDToSpan(span, userID)
		attachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)

		// fetch oauth2 client
		x, err := s.database.GetOAuth2Client(ctx, oauth2ClientID, userID)
		if err == sql.ErrNoRows {
			logger.Debug("ReadHandler called on nonexistent client")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error fetching oauth2Client from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode response and peace
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// ArchiveHandler is a route handler for archiving an OAuth2 client
func (s *Service) ArchiveHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ArchiveHandler")
		defer span.End()

		// determine subject matter
		userID := s.fetchUserID(req)
		oauth2ClientID := s.urlClientIDExtractor(req)

		logger := s.logger.WithValues(map[string]interface{}{
			"oauth2_client_id": oauth2ClientID,
			"user_id":          userID,
		})
		attachUserIDToSpan(span, userID)
		attachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)

		// mark client as archived
		err := s.database.ArchiveOAuth2Client(ctx, oauth2ClientID, userID)
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "encountered error deleting oauth2 client")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// notify relevant parties
		s.oauth2ClientCounter.Decrement(ctx)
		res.WriteHeader(http.StatusNoContent)
	}
}
