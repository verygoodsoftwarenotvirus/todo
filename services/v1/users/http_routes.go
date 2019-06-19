package users

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"strconv"

	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
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

func attachUsernameToSpan(span *trace.Span, username string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("username", username))
	}
}

func attachUserIDToSpan(span *trace.Span, userID uint64) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))
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

func (s *Service) validateCredentialChangeRequest(
	ctx context.Context,
	userID uint64,
	password,
	totpToken string,
) (user *models.User, httpStatus int) {
	ctx, span := trace.StartSpan(ctx, "validateCredentialChangeRequest")
	defer span.End()

	logger := s.logger.WithValue("user_id", userID)

	user, err := s.database.GetUser(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, http.StatusNotFound
	} else if err != nil {
		logger.Error(err, "error encountered fetching user")
		return nil, http.StatusInternalServerError
	}

	logger = s.logger.WithValue("user", user.ID)

	valid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		password,
		user.TwoFactorSecret,
		totpToken,
		user.Salt,
	)

	if err != nil {
		logger.Error(err, "error encountered generating random TOTP string")
		return nil, http.StatusInternalServerError
	} else if !valid {
		logger.WithValue("valid", valid).Error(err, "invalid attempt to cycle TOTP token")
		return nil, http.StatusUnauthorized
	}

	return user, http.StatusOK
}

// ListHandler is a handler for responding with a list of users
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "ListHandler")
	defer span.End()

	qf := models.ExtractQueryFilter(req)
	users, err := s.database.GetUsers(ctx, qf)
	if err != nil {
		s.logger.Error(err, "error fetching users for ListHandler route")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.encoderDecoder.EncodeResponse(res, users); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// CreateHandler is our user creation route
func (s *Service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "CreateHandler")
	defer span.End()

	if !s.userCreationEnabled {
		s.logger.Info("disallowing user creation")
		res.WriteHeader(http.StatusForbidden)
		return
	}

	input, ok := ctx.Value(UserCreationMiddlewareCtxKey).(*models.UserInput)
	if !ok {
		s.logger.Info("valid input not attached to UsersService CreateHandler request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	attachUsernameToSpan(span, input.Username)

	// NOTE: I feel comfortable letting username be in the logger, since
	// the logging statements below are only in the event of errors. If
	// and when that changes, this can/should be removed.
	logger := s.logger.WithValue("username", input.Username)

	hp, err := s.authenticator.HashPassword(ctx, input.Password)
	if err != nil {
		logger.Error(err, "valid input not attached to request")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	input.Password = hp

	input.TwoFactorSecret, err = randString()
	if err != nil {
		logger.Error(err, "error generating TOTP secret")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := s.database.CreateUser(ctx, input)
	if err != nil {
		if err == dbclient.ErrUserExists {
			logger.Info("duplicate username attempted")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		logger.Error(err, "error creating user")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// UserCreationResponse is a struct we can use to notify the user of
	// their two factor secret, but ideally just this once and then never again.
	ucr := &models.UserCreationResponse{
		ID:                    user.ID,
		Username:              user.Username,
		TwoFactorSecret:       user.TwoFactorSecret,
		PasswordLastChangedOn: user.PasswordLastChangedOn,
		CreatedOn:             user.CreatedOn,
		UpdatedOn:             user.UpdatedOn,
		ArchivedOn:            user.ArchivedOn,
		TwoFactorQRCode:       s.buildQRCode(ctx, user),
	}

	attachUserIDToSpan(span, user.ID)
	s.userCounter.Increment(ctx)
	s.reporter.Report(newsman.Event{
		EventType: string(models.Create),
		Data:      ucr,
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusCreated)
	if err = s.encoderDecoder.EncodeResponse(res, ucr); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

func (s *Service) buildQRCode(ctx context.Context, user *models.User) string {
	_, span := trace.StartSpan(ctx, "buildQRCode")
	defer span.End()

	// "otpauth://totp/{{ .Issuer }}:{{ .Username }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
	qrcode, err := qr.Encode(
		fmt.Sprintf(
			"otpauth://totp/%s:%s?secret=%s&issuer=%s",
			"todoservice",
			user.Username,
			user.TwoFactorSecret,
			"todoService",
		), qr.L, qr.Auto,
	)
	if err != nil {
		s.logger.Error(err, "trying to encode secret to qr code")
		return ""
	}

	qrcode, err = barcode.Scale(qrcode, 256, 256)
	if err != nil {
		s.logger.Error(err, "trying to enlarge qr code")
		return ""
	}

	var b bytes.Buffer
	if err = png.Encode(&b, qrcode); err != nil {
		s.logger.Error(err, "trying to encode qr code to png")
		return ""
	}

	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(b.Bytes()))
}

// ReadHandler is our read route
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "ReadHandler")
	defer span.End()

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user_id", userID)

	attachUserIDToSpan(span, userID)

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

	if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// NewTOTPSecretHandler fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid
func (s *Service) NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "NewTOTPSecretHandler")
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

	user, sc := s.validateCredentialChangeRequest(
		ctx,
		userID,
		input.CurrentPassword,
		input.TOTPToken,
	)

	attachUserIDToSpan(span, userID)

	if sc != http.StatusOK {
		res.WriteHeader(sc)
		return
	}

	attachUsernameToSpan(span, user.Username)
	logger := s.logger.WithValue("user", user.ID)

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

	res.WriteHeader(http.StatusAccepted)
	if err := s.encoderDecoder.EncodeResponse(
		res, &models.TOTPSecretRefreshResponse{TwoFactorSecret: user.TwoFactorSecret},
	); err != nil {
		s.logger.Error(err, "encoding response")
	}
}

// UpdatePasswordHandler updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid
func (s *Service) UpdatePasswordHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "UpdatePasswordHandler")
	defer span.End()

	input, ok := ctx.Value(PasswordChangeMiddlewareCtxKey).(*models.PasswordUpdateInput)
	if !ok {
		s.logger.Debug("no input found on UpdatePasswordHandler request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := ctx.Value(models.UserIDKey).(uint64)
	if !ok {
		s.logger.Debug("no user ID attached to UpdatePasswordHandler request")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, sc := s.validateCredentialChangeRequest(
		ctx,
		userID,
		input.CurrentPassword,
		input.TOTPToken,
	)

	attachUserIDToSpan(span, userID)

	if sc != http.StatusOK {
		res.WriteHeader(sc)
		return
	}

	attachUsernameToSpan(span, user.Username)
	logger := s.logger.WithValue("user", user.ID)

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

// ArchiveHandler is a handler for archiving a user
func (s *Service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := trace.StartSpan(req.Context(), "ArchiveHandler")
	defer span.End()

	userID := s.userIDFetcher(req)
	logger := s.logger.WithValue("user_id", userID)
	attachUserIDToSpan(span, userID)

	if err := s.database.ArchiveUser(ctx, userID); err != nil {
		logger.Error(err, "deleting user from database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.userCounter.Decrement(ctx)

	s.reporter.Report(newsman.Event{
		EventType: string(models.Archive),
		Data:      models.User{ID: userID},
		Topics:    []string{topicName},
	})

	res.WriteHeader(http.StatusNoContent)
}
