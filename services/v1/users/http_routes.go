package users

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"

	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/pquerna/otp/totp"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

const (
	// URIParamKey is used to refer to user IDs in router params.
	URIParamKey = "userID"

	totpIssuer = "todoService"
)

// validateCredentialChangeRequest takes a user's credentials and determines
// if they match what is on record.
func (s *Service) validateCredentialChangeRequest(
	ctx context.Context,
	userID uint64,
	password,
	totpToken string,
) (user *models.User, httpStatus int) {
	ctx, span := tracing.StartSpan(ctx, "validateCredentialChangeRequest")
	defer span.End()

	logger := s.logger.WithValue("user_id", userID)

	// fetch user data.
	user, err := s.userDataManager.GetUser(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, http.StatusNotFound
	} else if err != nil {
		logger.Error(err, "error encountered fetching user")
		return nil, http.StatusInternalServerError
	}

	// validate login.
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

// ListHandler is a handler for responding with a list of users.
func (s *Service) ListHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "ListHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// determine desired filter.
		qf := models.ExtractQueryFilter(req)

		// fetch user data.
		users, err := s.userDataManager.GetUsers(ctx, qf)
		if err != nil {
			logger.Error(err, "error fetching users for ListHandler route")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode response.
		if err = s.encoderDecoder.EncodeResponse(res, users); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// CreateHandler is our user creation route.
func (s *Service) CreateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "CreateHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// in the event that we don't want new users to be able to sign up (a config setting)
		// just decline the request from the get-go
		if !s.userCreationEnabled {
			logger.Info("disallowing user creation")
			res.WriteHeader(http.StatusForbidden)
			return
		}

		// fetch parsed input from request context.
		userInput, ok := ctx.Value(UserCreationMiddlewareCtxKey).(*models.UserCreationInput)
		if !ok {
			logger.Info("valid input not attached to UsersService CreateHandler request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		tracing.AttachUsernameToSpan(span, userInput.Username)

		// NOTE: I feel comfortable letting username be in the logger, since
		// the logging statements below are only in the event of errors. If
		// and when that changes, this can/should be removed.
		logger = logger.WithValue("username", userInput.Username)

		// hash the password.
		hp, err := s.authenticator.HashPassword(ctx, userInput.Password)
		if err != nil {
			logger.Error(err, "valid input not attached to request")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		input := models.UserDatabaseCreationInput{
			Username:        userInput.Username,
			HashedPassword:  hp,
			TwoFactorSecret: "",
		}

		// generate a two factor secret.
		input.TwoFactorSecret, err = s.secretGenerator.GenerateTwoFactorSecret()
		if err != nil {
			logger.Error(err, "error generating TOTP secret")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// create the user.
		user, err := s.userDataManager.CreateUser(ctx, input)
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

		// notify the relevant parties.
		tracing.AttachUserIDToSpan(span, user.ID)
		s.userCounter.Increment(ctx)
		s.reporter.Report(newsman.Event{
			EventType: string(models.Create),
			Data:      ucr,
			Topics:    []string{topicName},
		})

		// encode and peace.
		res.WriteHeader(http.StatusCreated)
		if err = s.encoderDecoder.EncodeResponse(res, ucr); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// buildQRCode builds a QR code for a given username and secret.
func (s *Service) buildQRCode(ctx context.Context, username, twoFactorSecret string) string {
	_, span := tracing.StartSpan(ctx, "buildQRCode")
	defer span.End()

	// encode two factor secret as authenticator-friendly QR code
	qrcode, err := qr.Encode(
		// "otpauth://totp/{{ .Issuer }}:{{ .Username }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
		fmt.Sprintf(
			"otpauth://totp/%s:%s?secret=%s&issuer=%s",
			totpIssuer,
			username,
			twoFactorSecret,
			totpIssuer,
		),
		qr.L,
		qr.Auto,
	)
	if err != nil {
		s.logger.Error(err, "trying to encode secret to qr code")
		return ""
	}

	// scale the QR code so that it's not a PNG for ants.
	qrcode, err = barcode.Scale(qrcode, 256, 256)
	if err != nil {
		s.logger.Error(err, "trying to enlarge qr code")
		return ""
	}

	// encode the QR code to PNG.
	var b bytes.Buffer
	if err = png.Encode(&b, qrcode); err != nil {
		s.logger.Error(err, "trying to encode qr code to png")
		return ""
	}

	// base64 encode the image for easy HTML use.
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(b.Bytes()))
}

// ReadHandler is our read route.
func (s *Service) ReadHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "ReadHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// figure out who this is all for.
		userID := s.userIDFetcher(req)
		logger = logger.WithValue("user_id", userID)

		// document it for posterity.
		tracing.AttachUserIDToSpan(span, userID)

		// fetch user data.
		x, err := s.userDataManager.GetUser(ctx, userID)
		if err == sql.ErrNoRows {
			logger.Debug("no such user")
			res.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error(err, "error fetching user from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// encode response and peace.
		if err = s.encoderDecoder.EncodeResponse(res, x); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// TOTPSecretVerificationHandler accepts a TOTP token as input and returns 200 if the TOTP token
// is validated by the user's TOTP secret.
func (s *Service) TOTPSecretVerificationHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "TOTPSecretVerificationHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// check request context for parsed input.
		input, ok := req.Context().Value(TOTPSecretVerificationMiddlewareCtxKey).(*models.TOTPSecretVerificationInput)
		if !ok || input == nil {
			logger.Debug("no input found on TOTP secret refresh request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := s.userDataManager.GetUserWithUnverifiedTwoFactorSecret(ctx, input.UserID)
		if err != nil {
			logger.Error(err, "fetching user")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		tracing.AttachUserIDToSpan(span, user.ID)
		tracing.AttachUsernameToSpan(span, user.Username)

		if user.TwoFactorSecretVerifiedOn != nil {
			// I suppose if this happens too many times, we'll want to keep track of that
			res.WriteHeader(http.StatusAlreadyReported)
			return
		}

		if totp.Validate(input.TOTPToken, user.TwoFactorSecret) {
			user.TwoFactorSecretVerifiedOn = nil
			if updateUserErr := s.userDataManager.VerifyUserTwoFactorSecret(ctx, user.ID); updateUserErr != nil {
				logger.Error(updateUserErr, "updating user to indicate their 2FA secret is validated")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			res.WriteHeader(http.StatusAccepted)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	}
}

// NewTOTPSecretHandler fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid.
func (s *Service) NewTOTPSecretHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "NewTOTPSecretHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// check request context for parsed input.
		input, ok := req.Context().Value(TOTPSecretRefreshMiddlewareCtxKey).(*models.TOTPSecretRefreshInput)
		if !ok {
			logger.Debug("no input found on TOTP secret refresh request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// also check for the user's ID.
		userID, ok := ctx.Value(models.UserIDKey).(uint64)
		if !ok {
			logger.Debug("no user ID attached to TOTP secret refresh request")
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

		// if the above function returns something other than 200, it means some error occurred.
		if sc != http.StatusOK {
			res.WriteHeader(sc)
			return
		}

		// document who this is for.
		tracing.AttachUserIDToSpan(span, userID)
		tracing.AttachUsernameToSpan(span, user.Username)
		logger = logger.WithValue("user", user.ID)

		// set the two factor secret.
		tfs, err := s.secretGenerator.GenerateTwoFactorSecret()
		if err != nil {
			logger.Error(err, "error encountered generating random TOTP string")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.TwoFactorSecret = tfs
		user.TwoFactorSecretVerifiedOn = nil

		// update the user in the database.
		if err := s.userDataManager.UpdateUser(ctx, user); err != nil {
			logger.Error(err, "error encountered updating TOTP token")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// let the requester know we're all good.
		res.WriteHeader(http.StatusAccepted)
		if err := s.encoderDecoder.EncodeResponse(res, &models.TOTPSecretRefreshResponse{TwoFactorSecret: user.TwoFactorSecret}); err != nil {
			logger.Error(err, "encoding response")
		}
	}
}

// UpdatePasswordHandler updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid.
func (s *Service) UpdatePasswordHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "UpdatePasswordHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// check request context for parsed value.
		input, ok := ctx.Value(PasswordChangeMiddlewareCtxKey).(*models.PasswordUpdateInput)
		if !ok {
			logger.Debug("no input found on UpdatePasswordHandler request")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// check request context for user ID.
		userID, ok := ctx.Value(models.UserIDKey).(uint64)
		if !ok {
			logger.Debug("no user ID attached to UpdatePasswordHandler request")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		// determine relevant user ID.
		tracing.AttachUserIDToSpan(span, userID)
		logger = logger.WithValue("user_id", userID)

		// make sure everything's on the up-and-up
		user, sc := s.validateCredentialChangeRequest(
			ctx,
			userID,
			input.CurrentPassword,
			input.TOTPToken,
		)

		// if the above function returns something other than 200, it means some error occurred.
		if sc != http.StatusOK {
			res.WriteHeader(sc)
			return
		}

		tracing.AttachUsernameToSpan(span, user.Username)

		// hash the new password.
		var err error
		user.HashedPassword, err = s.authenticator.HashPassword(ctx, input.NewPassword)
		if err != nil {
			logger.Error(err, "error hashing password")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// update the user.
		if err = s.userDataManager.UpdateUser(ctx, user); err != nil {
			logger.Error(err, "error encountered updating user")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// we're all good.
		res.WriteHeader(http.StatusAccepted)
	}
}

// ArchiveHandler is a handler for archiving a user.
func (s *Service) ArchiveHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "ArchiveHandler")
		defer span.End()

		logger := s.logger.WithRequest(req)

		// figure out who this is for.
		userID := s.userIDFetcher(req)
		logger = logger.WithValue("user_id", userID)
		tracing.AttachUserIDToSpan(span, userID)

		// do the deed.
		if err := s.userDataManager.ArchiveUser(ctx, userID); err != nil {
			logger.Error(err, "deleting user from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// inform the relatives.
		s.userCounter.Decrement(ctx)
		s.reporter.Report(newsman.Event{
			EventType: string(models.Archive),
			Data:      models.User{ID: userID},
			Topics:    []string{topicName},
		})

		// we're all good.
		res.WriteHeader(http.StatusNoContent)
	}
}
