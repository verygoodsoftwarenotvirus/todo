package users

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"strconv"

	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	v1 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/events/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"go.opencensus.io/trace"
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

func (s *Service) validateCredentialChangeRequest(req *http.Request, userID uint64, password, totpToken string) (*models.User, int) {
	logger := s.logger.WithValue("user_id", userID)

	ctx := req.Context()
	user, err := s.database.GetUser(ctx, userID)
	if err != nil {
		logger.Error(err, "error encountered fetching user")
		return nil, http.StatusInternalServerError
	}

	logger = s.logger.WithValue("username", user.Username)

	valid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		user.Salt,
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

// List is a handler for responding with a list of users
func (s *Service) List(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "list_route")
	defer span.End()

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

// Create is our user creation route
func (s *Service) Create(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "create_route")
	defer span.End()

	input, ok := ctx.Value(UserCreationMiddlewareCtxKey).(*models.UserInput)
	if !ok {
		s.logger.Error(nil, "valid input not attached to UsersService Create request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	logger := s.logger.WithValues(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	})
	logger.Debug("user creation route hit")

	span.AddAttributes(
		trace.StringAttribute("username", input.Username),
		trace.BoolAttribute("is_admin", input.IsAdmin),
	)

	hp, err := s.authenticator.HashPassword(ctx, input.Password)
	if err != nil {
		logger.Error(err, "valid input not attached to request")
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
		if err == dbclient.ErrUserExists {
			s.logger.Debug("duplicate username attempted")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		s.logger.Error(err, "error creating user")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// "otpauth://totp/{{ .Issuer }}:{{ .Username }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
	qrcode, err := qr.Encode(
		fmt.Sprintf(
			"otpauth://totp/%s:%s?secret=%s&issuer=%s",
			"todoservice",
			user.Username,
			user.TwoFactorSecret,
			"todoService",
		),
		qr.L,
		qr.Auto,
	)
	if err != nil {
		s.logger.Error(err, "trying to encode secret to qr code")
	}
	qrcode, err = barcode.Scale(qrcode, 256, 256)
	if err != nil {
		s.logger.Error(err, "trying to enlarge qr code")
	}

	var b bytes.Buffer
	if err := png.Encode(&b, qrcode); err != nil {
		s.logger.Error(err, "trying to encode qr code to png")
	}

	qrCode := fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(b.Bytes()))

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
		TwoFactorQRCode:       qrCode,
	}

	s.userCounter.Increment(ctx)

	s.newsman.Report(newsman.Event{
		EventType: string(v1.Create),
		Data:      x,
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusCreated)
	if err = s.encoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// Read is our read route
func (s *Service) Read(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "read_route")
	defer span.End()

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user_id", userID)
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

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

// NewTOTPSecret fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid
func (s *Service) NewTOTPSecret(res http.ResponseWriter, req *http.Request) {
	s.logger.Debug("NewTOTPSecret route hit in UsersService")
	ctx, span := trace.StartSpan(req.Context(), "new_totp_secret")
	defer span.End()

	var err error
	input, ok := req.Context().Value(TOTPSecretRefreshMiddlewareCtxKey).(*models.TOTPSecretRefreshInput)
	if !ok {
		s.logger.Debug("no input found on TOTP secret refresh request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := ctx.Value(models.UserIDKey).(uint64)
	if !ok {
		s.logger.Debug("no user ID attached to TOTP secret refresh request")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, sc := s.validateCredentialChangeRequest(req, userID, input.CurrentPassword, input.TOTPToken)
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

	logger.WithValue("generated_totp_secret", user.TwoFactorSecret)

	if err := s.database.UpdateUser(ctx, user); err != nil {
		logger.Error(err, "error encountered updating TOTP token")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	if err := s.encoder.EncodeResponse(
		res, &models.TOTPSecretRefreshResponse{TwoFactorSecret: user.TwoFactorSecret},
	); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// UpdatePassword updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid
func (s *Service) UpdatePassword(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "update_password_route")
	defer span.End()

	input, ok := ctx.Value(PasswordChangeMiddlewareCtxKey).(*models.PasswordUpdateInput)
	if !ok {
		s.logger.Debug("no input found on UpdatePassword request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := ctx.Value(models.UserIDKey).(uint64)
	if !ok {
		s.logger.Debug("no user ID attached to UpdatePassword request")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, sc := s.validateCredentialChangeRequest(req, userID, input.CurrentPassword, input.TOTPToken)
	if sc != 0 {
		res.WriteHeader(sc)
		return
	}

	logger := s.logger.WithValue("username", user.Username)

	var err error
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
}

// Delete is a handler for deleting a user
func (s *Service) Delete(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "delete_route")
	defer span.End()

	userID := s.userIDFetcher(req)

	logger := s.logger.WithValue("user_id", userID)
	logger.Debug("UsersService.Delete called")
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	if err := s.database.DeleteUser(ctx, userID); err != nil {
		logger.Error(err, "UsersService.Delete called")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.userCounter.Decrement(ctx)

	s.newsman.Report(newsman.Event{
		EventType: string(v1.Delete),
		Data:      models.User{ID: userID},
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusNoContent)
}
