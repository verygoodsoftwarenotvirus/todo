package users

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const (
	// URIParamKey is used to refer to user IDs in router params
	URIParamKey = "userID"
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

// UserLoginInputContextMiddleware fetches user login input from requests
func (s *Service) UserLoginInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserLoginInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithRequest(req).Debug("UserLoginInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// UserInputContextMiddleware fetches user input from requests
func (s *Service) UserInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.UserInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithRequest(req).Debug("UserInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// PasswordUpdateInputContextMiddleware fetches password update input from requests
func (s *Service) PasswordUpdateInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.PasswordUpdateInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithRequest(req).Debug("PasswordUpdateInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// TOTPSecretRefreshInputContextMiddleware fetches 2FA update input from requests
func (s *Service) TOTPSecretRefreshInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.TOTPSecretRefreshInput)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithRequest(req).Debug("TOTPSecretRefreshInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// Read is our read route
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("Read route hit in UsersService")

	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("read route", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	userID := s.usernameFetcher(req)
	logger := s.logger.WithValue("user_id", userID)

	x, err := s.database.GetUser(ctx, userID)
	if err == sql.ErrNoRows {
		logger.Debug("no such user")
		res.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		logger.Error(err, "error fetching user from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Count is a handler for responding with a count of users
func (s *Service) Count(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("Count route hit in UsersService")

	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("count route", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	qf := models.ExtractQueryFilter(req)
	logger := s.logger.WithValue("query_filter", qf)

	userCount, err := s.database.GetUserCount(ctx, qf)
	if err != nil {
		logger.Error(err, "error fetching item count from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	x := &models.CountResponse{Count: userCount}
	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// List is a handler for responding with a list of users
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("List route hit in UsersService")

	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("list route", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	qf := models.ExtractQueryFilter(req)
	logger := s.logger.WithValue("query_filter", qf)

	users, err := s.database.GetUsers(ctx, qf)
	if err != nil {
		logger.Error(err, "error fetching users for List route")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, users); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Delete is a handler for deleting a user
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("Delete route hit in UsersService")

	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("delete route", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	username := s.usernameFetcher(req)
	logger := s.logger.WithValue("username", username)
	logger.Debug("UsersService.Delete called")

	if err := s.database.DeleteUser(ctx, username); err != nil {
		logger.Error(err, "UsersService.Delete called")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusNoContent)
}

func (s *Service) validateCredentialChangeRequest(req *http.Request, password string, totpToken string) (*models.User, int) {
	username := s.usernameFetcher(req)
	logger := s.logger.WithValue("username", username)

	ctx := req.Context()
	user, err := s.database.GetUser(ctx, username)
	if err != nil {
		logger.Error(err, "error encountered fecthing user")
		return nil, http.StatusInternalServerError
	}

	logger = logger.WithValue("username", user.Username)

	valid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		password,
		user.TwoFactorSecret,
		totpToken,
	)

	if err != nil {
		logger.Error(err, "error encountered generating random TOTP string")
		return nil, http.StatusInternalServerError
	} else if !valid {
		logger.WithValue("valid", valid).Error(err, "invalid attempt to cycle TOTP token")
		return nil, http.StatusUnauthorized
	}

	return user, 0
}

// NewTOTPSecret fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid
func (s *Service) NewTOTPSecret(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("NewTOTPSecret route hit in UsersService")

	var err error
	ctx := req.Context()
	input, ok := req.Context().Value(MiddlewareCtxKey).(*models.TOTPSecretRefreshInput)
	if !ok {
		s.logger.Debug("no input found on TOTP Secret refresh request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, sc := s.validateCredentialChangeRequest(req, input.CurrentPassword, input.TOTPToken)
	if sc != 0 {
		res.WriteHeader(sc)
		return
	}

	logger := s.logger.WithValue("username", user.Username)

	tfc, err := randString()
	if err != nil {
		logger.Error(err, "error encountered generating random TOTP string")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.TwoFactorSecret = tfc

	if err = s.database.UpdateUser(ctx, user); err != nil {
		logger.Error(err, "error encountered updating TOTP token")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoder.EncodeResponse(res, user); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// UpdatePassword updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid
func (s *Service) UpdatePassword(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("UpdatePassword route hit in UsersService")

	var err error
	ctx := req.Context()
	input, ok := ctx.Value(MiddlewareCtxKey).(*models.PasswordUpdateInput)
	if !ok {
		s.logger.Debug("no input found on TOTP Secret refresh request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, sc := s.validateCredentialChangeRequest(req, input.CurrentPassword, input.TOTPToken)
	if sc != 0 {
		res.WriteHeader(sc)
		return
	}

	logger := s.logger.WithValue("username", user.Username)

	user.HashedPassword, err = s.authenticator.HashPassword(ctx, input.NewPassword)
	if err != nil {
		logger.Error(err, "error hashing password")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.database.UpdateUser(ctx, user); err != nil {
		logger.Error(err, "error encountered updating user")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)
	//if err := s.encoder.EncodeResponse(res, user); err != nil {
	//	s.logger.Error(err, "encoding response")
	//}
}

// Create is our user creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	spanCtx, _ := s.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := s.tracer.StartSpan("create route", opentracing.ChildOf(spanCtx))
	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	defer serverSpan.Finish()

	input, ok := ctx.Value(MiddlewareCtxKey).(*models.UserInput)
	if !ok {
		s.logger.Error(nil, "valid input not attached to UsersService Create request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	s.logger.WithValues(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	}).Debug("user creation route hit")

	hp, err := s.authenticator.HashPassword(ctx, input.Password)
	if err != nil {
		s.logger.Error(err, "valid input not attached to request")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	input.Password = hp

	input.TwoFactorSecret, err = randString()
	if err != nil {
		s.logger.Error(err, "error generating TOTP secret")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := s.database.CreateUser(ctx, input)
	if err != nil {
		s.logger.Error(err, "error creating user")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// UserCreationResponse is a struct we can use to notify the user of
	// their two factor secret, but ideally just this once and then never again.
	x := &models.UserCreationResponse{
		ID:                    user.ID,
		Username:              user.Username,
		TwoFactorSecret:       user.TwoFactorSecret,
		PasswordLastChangedOn: user.PasswordLastChangedOn,
		CreatedOn:             user.CreatedOn,
		UpdatedOn:             user.UpdatedOn,
		ArchivedOn:            user.ArchivedOn,
	}

	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}
