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
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"go.opencensus.io/trace"
)

const (
	// URIParamKey is used to refer to user IDs in router params
	URIParamKey = "userID"
)

// this function tests that we have appropriate access to crypto/rand
func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

// attachUsernameToSpan provides a consistent way to attach a username to a span
func attachUsernameToSpan(span *trace.Span, username string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("username", username))
	}
}

// attachUserIDToSpan provides a consistent way to attach a user ID to a span
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

// validateCredentialChangeRequest takes a user's credentials and determines
// if they match what is on record
func (s *Service) validateCredentialChangeRequest(
	ctx context.Context,
	userID uint64,
	password,
	totpToken string,
) (user *models.User, httpStatus int) {
	ctx, span := trace.StartSpan(ctx, "validateCredentialChangeRequest")
	defer span.End()

	logger := s.logger.WithValue("user_id", userID)

	// fetch user data
	user, err := s.database.GetUser(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, http.StatusNotFound
	} else if err != nil {
		logger.Error(err, "error encountered fetching user")
		return nil, http.StatusInternalServerError
	}

	// validate login
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
func (s *Service) ListHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ListHandler")
		defer span.End()

		// determine desired filter
		qf := models.ExtractQueryFilter(req)

		// fetch user data
		users, err := s.database.GetUsers(ctx, qf)
		if err != nil {
			s.logger.Error(err, "error fetching users for ListHandler route")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode response
		if err = s.encoderDecoder.EncodeResponse(res, users); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// CreateHandler is our user creation route
func (s *Service) CreateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "CreateHandler")
		defer span.End()

		// in the event that we don't want new users to be able to sign up (a config setting)
		// just decline the request from the get-go
		if !s.userCreationEnabled {
			s.logger.Info("disallowing user creation")
			res.WriteHeader(http.StatusForbidden)
			return
		}

		// fetch parsed input from request context
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

		// hash the password
		hp, err := s.authenticator.HashPassword(ctx, input.Password)
		if err != nil {
			logger.Error(err, "valid input not attached to request")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		input.Password = hp

		// generate a two factor secret
		input.TwoFactorSecret, err = randString()
		if err != nil {
			logger.Error(err, "error generating TOTP secret")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// create the user
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
			TwoFactorQRCode:       s.buildQRCode(ctx, user.Username, user.TwoFactorSecret),
		}

		// notify the relevant parties
		attachUserIDToSpan(span, user.ID)
		s.userCounter.Increment(ctx)
		s.reporter.Report(newsman.Event{
			EventType: string(models.Create),
			Data:      ucr,
			Topics:    []string{topicName},
		})

		// encode and peace
		res.WriteHeader(http.StatusCreated)
		if err = s.encoderDecoder.EncodeResponse(res, ucr); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// buildQRCode builds a QR code for a given username and secret
func (s *Service) buildQRCode(ctx context.Context, username, twoFactorSecret string) string {
	_, span := trace.StartSpan(ctx, "buildQRCode")
	defer span.End()

	// encode two factor secret as authenticator-friendly QR code
	qrcode, err := qr.Encode(
		// "otpauth://totp/{{ .Issuer }}:{{ .Username }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
		fmt.Sprintf(
			"otpauth://totp/%s:%s?secret=%s&issuer=%s",
			"todoservice",
			username,
			twoFactorSecret,
			"todoService",
		),
		qr.L,
		qr.Auto,
	)
	if err != nil {
		s.logger.Error(err, "trying to encode secret to qr code")
		return ""
	}

	// scale the QR code so that it's not a PNG for ants
	qrcode, err = barcode.Scale(qrcode, 256, 256)
	if err != nil {
		s.logger.Error(err, "trying to enlarge qr code")
		return ""
	}

	// encode the QR code to PNG
	var b bytes.Buffer
	if err = png.Encode(&b, qrcode); err != nil {
		s.logger.Error(err, "trying to encode qr code to png")
		return ""
	}

	// base64 encode the image for easy HTML use
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(b.Bytes()))
}

// ReadHandler is our read route
func (s *Service) ReadHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ReadHandler")
		defer span.End()

		// figure out who this is all for
		userID := s.userIDFetcher(req)
		logger := s.logger.WithValue("user_id", userID)

		// document it for posterity
		attachUserIDToSpan(span, userID)

		// fetch user data
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

		// encode response and peace
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// NewTOTPSecretHandler fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid
func (s *Service) NewTOTPSecretHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "NewTOTPSecretHandler")
		defer span.End()

		// check request context for parsed input
		input, ok := req.Context().Value(TOTPSecretRefreshMiddlewareCtxKey).(*models.TOTPSecretRefreshInput)
		if !ok {
			s.logger.Debug("no input found on TOTP secret refresh request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// also check for the user's ID
		userID, ok := ctx.Value(models.UserIDKey).(uint64)
		if !ok {
			s.logger.Debug("no user ID attached to TOTP secret refresh request")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		// make sure this is all on the up-and-up
		user, sc := s.validateCredentialChangeRequest(
			ctx,
			userID,
			input.CurrentPassword,
			input.TOTPToken,
		)

		// if the above function returns something other than 200, it means some error occurred
		if sc != http.StatusOK {
			res.WriteHeader(sc)
			return
		}

		// document who this is for
		attachUserIDToSpan(span, userID)
		attachUsernameToSpan(span, user.Username)
		logger := s.logger.WithValue("user", user.ID)

		// set the two factor secret
		tfs, err := randString()
		if err != nil {
			logger.Error(err, "error encountered generating random TOTP string")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.TwoFactorSecret = tfs

		// update the user in the database
		if err := s.database.UpdateUser(ctx, user); err != nil {
			logger.Error(err, "error encountered updating TOTP token")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// let the requester know we're all good
		res.WriteHeader(http.StatusAccepted)
		if err := s.encoderDecoder.EncodeResponse(res, &models.TOTPSecretRefreshResponse{TwoFactorSecret: user.TwoFactorSecret}); err != nil {
			s.logger.Error(err, "encoding response")
		}
	}
}

// UpdatePasswordHandler updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid
func (s *Service) UpdatePasswordHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "UpdatePasswordHandler")
		defer span.End()

		// check request context for parsed value
		input, ok := ctx.Value(PasswordChangeMiddlewareCtxKey).(*models.PasswordUpdateInput)
		if !ok {
			s.logger.Debug("no input found on UpdatePasswordHandler request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// check request context for user ID
		userID, ok := ctx.Value(models.UserIDKey).(uint64)
		if !ok {
			s.logger.Debug("no user ID attached to UpdatePasswordHandler request")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		// make sure everything's on the up-and-up
		user, sc := s.validateCredentialChangeRequest(
			ctx,
			userID,
			input.CurrentPassword,
			input.TOTPToken,
		)

		// if the above function returns something other than 200, it means some error occurred
		if sc != http.StatusOK {
			res.WriteHeader(sc)
			return
		}

		// document who this is all for
		attachUserIDToSpan(span, userID)
		attachUsernameToSpan(span, user.Username)
		logger := s.logger.WithValue("user", user.ID)

		// hash the new password
		var err error
		user.HashedPassword, err = s.authenticator.HashPassword(ctx, input.NewPassword)
		if err != nil {
			logger.Error(err, "error hashing password")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// update the user
		if err = s.database.UpdateUser(ctx, user); err != nil {
			logger.Error(err, "error encountered updating user")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// we're all good
		res.WriteHeader(http.StatusAccepted)
	}
}

// ArchiveHandler is a handler for archiving a user
func (s *Service) ArchiveHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "ArchiveHandler")
		defer span.End()

		// figure out who this is for
		userID := s.userIDFetcher(req)
		logger := s.logger.WithValue("user_id", userID)
		attachUserIDToSpan(span, userID)

		// do the deed
		if err := s.database.ArchiveUser(ctx, userID); err != nil {
			logger.Error(err, "deleting user from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// inform the relatives
		s.userCounter.Decrement(ctx)
		s.reporter.Report(newsman.Event{
			EventType: string(models.Archive),
			Data:      models.User{ID: userID},
			Topics:    []string{topicName},
		})

		// we're all good
		res.WriteHeader(http.StatusNoContent)
	}
}
