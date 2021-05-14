package users

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"net/http"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp/totp"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	// UserIDURIParamKey is used to refer to user IDs in router params.
	UserIDURIParamKey = "userID"

	totpIssuer             = "todoService"
	base64ImagePrefix      = "data:image/jpeg;base64,"
	minimumPasswordEntropy = 75
	totpSecretSize         = 64
)

// validateCredentialChangeRequest takes a user's credentials and determines
// if they match what is on record.
func (s *service) validateCredentialChangeRequest(ctx context.Context, userID uint64, password, totpToken string) (user *types.User, httpStatus int) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.UserIDKey, userID)

	// fetch user data.
	user, err := s.userDataManager.GetUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, http.StatusNotFound
	} else if err != nil {
		logger.Error(err, "error encountered fetching user")
		return nil, http.StatusInternalServerError
	}

	// validate login.
	valid, validationErr := s.authenticator.ValidateLogin(ctx, user.HashedPassword, password, user.TwoFactorSecret, totpToken)
	if validationErr != nil {
		logger.WithValue("validation_error", validationErr).Debug("error validating credentials")
		return nil, http.StatusBadRequest
	} else if !valid {
		logger.WithValue("valid", valid).Error(err, "invalid credentials")
		return nil, http.StatusUnauthorized
	}

	return user, http.StatusOK
}

// UsernameSearchHandler is a handler for responding to username queries.
func (s *service) UsernameSearchHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	query := req.URL.Query().Get(types.SearchQueryKey)
	logger := s.logger.WithRequest(req).WithValue("query", query)

	// fetch user data.
	users, err := s.userDataManager.SearchForUsersByUsername(ctx, query)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "searching for users")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response.
	s.encoderDecoder.RespondWithData(ctx, res, users)
}

// ListHandler is a handler for responding with a list of users.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine desired filter.
	qf := types.ExtractQueryFilter(req)

	// fetch user data.
	users, err := s.userDataManager.GetUsers(ctx, qf)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching users")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response.
	s.encoderDecoder.RespondWithData(ctx, res, users)
}

func (s *service) RegisterUser(ctx context.Context, registrationInput *types.UserRegistrationInput) (*types.UserCreationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	// NOTE: I feel comfortable letting username be in the logger, since
	// the logging statements below are only in the event of errs. If
	// and when that changes, this can/should be removed.
	logger := s.logger.WithValue(keys.UsernameKey, registrationInput.Username)
	tracing.AttachUsernameToSpan(span, registrationInput.Username)

	// hash the password
	hp, err := s.authenticator.HashPassword(ctx, registrationInput.Password)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "hashing password")
	}

	input := &types.UserDataStoreCreationInput{
		Username:        registrationInput.Username,
		HashedPassword:  hp,
		TwoFactorSecret: "",
	}

	// generate a two factor secret.
	if input.TwoFactorSecret, err = s.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize); err != nil {
		return nil, observability.PrepareError(err, logger, span, "generating TOTP secret")
	}

	// create the user.
	user, userCreationErr := s.userDataManager.CreateUser(ctx, input)
	if userCreationErr != nil {
		return nil, observability.PrepareError(err, logger, span, "creating user")
	}

	// notify the relevant parties.
	tracing.AttachUserIDToSpan(span, user.ID)
	s.userCounter.Increment(ctx)

	// UserCreationResponse is a struct we can use to notify the user of their two factor secret, but ideally just this once and then never again.
	ucr := &types.UserCreationResponse{
		CreatedUserID:   user.ID,
		Username:        user.Username,
		CreatedOn:       user.CreatedOn,
		TwoFactorQRCode: s.buildQRCode(ctx, user.Username, user.TwoFactorSecret),
	}

	return ucr, nil
}

// CreateHandler is our user creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// in the event that we don't want new users to be able to sign up (a config setting)
	// just decline the request from the get-go
	if !s.authSettings.EnableUserSignup {
		logger.Info("disallowing user creation")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "user creation is disabled", http.StatusForbidden)
		return
	}

	// fetch parsed input from session context data.
	userInput, ok := ctx.Value(types.UserRegistrationInputContextKey).(*types.UserRegistrationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// NOTE: I feel comfortable letting username be in the logger, since
	// the logging statements below are only in the event of errs. If
	// and when that changes, this can/should be removed.
	logger = logger.WithValue(keys.UsernameKey, userInput.Username)
	tracing.AttachUsernameToSpan(span, userInput.Username)

	// ensure the password is not garbage-tier
	if err := passwordvalidator.Validate(userInput.Password, minimumPasswordEntropy); err != nil {
		logger.WithValue("password_validation_error", err).Debug("weak password provided to user creation route")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "password too weak", http.StatusBadRequest)
		return
	}

	ucr, err := s.RegisterUser(ctx, userInput)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode and peace.
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, ucr, http.StatusCreated)
}

// buildQRCode builds a QR code for a given username and secret.
func (s *service) buildQRCode(ctx context.Context, username, twoFactorSecret string) string {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.UsernameKey, username)

	// "otpauth://totp/{{ .Issuer }}:{{ .EnsureUsername }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
	otpString := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		totpIssuer,
		username,
		twoFactorSecret,
		totpIssuer,
	)

	x, err := qrcode.NewQRCodeWriter().EncodeWithoutHint(otpString, gozxing.BarcodeFormat_QR_CODE, 128, 128)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding secret to QR code")
		return ""
	}

	// encode the QR code to PNG.
	var b bytes.Buffer
	if err = png.Encode(&b, x); err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding QR code to PNG")
		return ""
	}

	// base64 encode the image for easy HTML use.
	return fmt.Sprintf("%s%s", base64ImagePrefix, base64.StdEncoding.EncodeToString(b.Bytes()))
}

// SelfHandler returns information about the user making the request.
func (s *service) SelfHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	// figure out who this is all for.
	requester := sessionCtxData.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestingUserIDToSpan(span, requester)

	// fetch user data.
	user, err := s.userDataManager.GetUser(ctx, requester)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no such user")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, user)
}

// ReadHandler is our read route.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out who this is all for.
	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	// fetch user data.
	x, err := s.userDataManager.GetUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no such user")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching user from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}

var errSecretAlreadyVerified = errors.New("secret already verified")

func (s *service) VerifyUserTwoFactorSecret(ctx context.Context, input *types.TOTPSecretVerificationInput) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.UserIDKey, input.UserID)

	user, err := s.userDataManager.GetUserWithUnverifiedTwoFactorSecret(ctx, input.UserID)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching user with unverified two factor secret")
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachUsernameToSpan(span, user.Username)

	if user.TwoFactorSecretVerifiedOn != nil {
		// I suppose if this happens too many times, we might want to keep track of that
		logger.Debug("two factor secret already verified")
		return errSecretAlreadyVerified
	}

	if totp.Validate(input.TOTPToken, user.TwoFactorSecret) {
		if updateUserErr := s.userDataManager.VerifyUserTwoFactorSecret(ctx, user.ID); updateUserErr != nil {
			return observability.PrepareError(err, logger, span, "marking 2FA secret as validated")
		}
	} else {
		return passwords.ErrInvalidTOTPToken
	}

	return nil
}

// TOTPSecretVerificationHandler accepts a TOTP token as input and returns 200 if the TOTP token
// is validated by the user's TOTP secret.
func (s *service) TOTPSecretVerificationHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check session context data for parsed input.
	input, ok := req.Context().Value(totpSecretVerificationMiddlewareCtxKey).(*types.TOTPSecretVerificationInput)
	if !ok || input == nil {
		logger.Debug("no input found on TOTP secret refresh request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.UserIDKey, input.UserID)

	if err := s.VerifyUserTwoFactorSecret(ctx, input); err != nil {
		switch {
		case errors.Is(err, passwords.ErrInvalidTOTPToken):
			s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		case errors.Is(err, errSecretAlreadyVerified):
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "TOTP secret already verified", http.StatusAlreadyReported)
			return
		default:
			observability.AcknowledgeError(err, logger, span, "verifying user two factor secret")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			return
		}
	}

	res.WriteHeader(http.StatusAccepted)
}

// NewTOTPSecretHandler fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid.
func (s *service) NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check session context data for parsed input.
	input, ok := req.Context().Value(totpSecretRefreshMiddlewareCtxKey).(*types.TOTPSecretRefreshInput)
	if !ok {
		logger.Debug("no input found on TOTP secret refresh request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	// make sure this is all on the up-and-up
	user, httpStatus := s.validateCredentialChangeRequest(
		ctx,
		sessionCtxData.Requester.ID,
		input.CurrentPassword,
		input.TOTPToken,
	)

	// if the above function returns something other than 200, it means some error occurred.
	if httpStatus != http.StatusOK {
		res.WriteHeader(httpStatus)
		return
	}

	// document who this is for.
	tracing.AttachRequestingUserIDToSpan(span, sessionCtxData.Requester.ID)
	tracing.AttachUsernameToSpan(span, user.Username)
	logger = logger.WithValue(keys.UserIDKey, user.ID)

	// set the two factor secret.
	tfs, err := s.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "generating 2FA secret")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	user.TwoFactorSecret = tfs
	user.TwoFactorSecretVerifiedOn = nil

	// update the user in the database.
	if err = s.userDataManager.UpdateUser(ctx, user, nil); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating 2FA secret")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// let the requester know we're all good.
	result := &types.TOTPSecretRefreshResponse{
		TwoFactorSecret: user.TwoFactorSecret,
		TwoFactorQRCode: s.buildQRCode(ctx, user.Username, user.TwoFactorSecret),
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, result, http.StatusAccepted)
}

// UpdatePasswordHandler updates a user's passwords, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid.
func (s *service) UpdatePasswordHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check session context data for parsed value.
	input, ok := ctx.Value(passwordChangeMiddlewareCtxKey).(*types.PasswordUpdateInput)
	if !ok {
		logger.Debug("no input found on UpdatePasswordHandler request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	// determine relevant user ID.
	tracing.AttachRequestingUserIDToSpan(span, sessionCtxData.Requester.ID)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	// make sure everything's on the up-and-up
	user, httpStatus := s.validateCredentialChangeRequest(
		ctx,
		sessionCtxData.Requester.ID,
		input.CurrentPassword,
		input.TOTPToken,
	)

	// if the above function returns something other than 200, it means some error occurred.
	if httpStatus != http.StatusOK {
		res.WriteHeader(httpStatus)
		return
	}

	tracing.AttachUsernameToSpan(span, user.Username)

	// ensure the passwords isn't garbage-tier
	if err = passwordvalidator.Validate(input.NewPassword, minimumPasswordEntropy); err != nil {
		logger.WithValue("password_validation_error", err).Debug("invalid password provided")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "new passwords is too weak!", http.StatusBadRequest)
		return
	}

	// hash the new passwords.
	newPasswordHash, err := s.authenticator.HashPassword(ctx, input.NewPassword)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "hashing password")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the user.
	if err = s.userDataManager.UpdateUserPassword(ctx, user.ID, newPasswordHash); err != nil {
		observability.AcknowledgeError(err, logger, span, "encountered updating user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// we're all good, log the user out
	http.SetCookie(res, &http.Cookie{MaxAge: -1})
}

func stringPointer(storageProviderPath string) *string {
	return &storageProviderPath
}

// AvatarUploadHandler updates a user's avatar.
func (s *service) AvatarUploadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)
	logger.Debug("session context data data extracted")

	user, err := s.userDataManager.GetUser(ctx, sessionCtxData.Requester.ID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching associated user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)
	logger.Debug("retrieved user from database")

	img, err := s.imageUploadProcessor.Process(ctx, req, "avatar")
	if err != nil || img == nil {
		observability.AcknowledgeError(err, logger, span, "processing provided avatar upload file")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	internalPath := fmt.Sprintf("avatar_%d", sessionCtxData.Requester.ID)
	logger = logger.WithValue("file_size", len(img.Data)).WithValue("internal_path", internalPath)

	if err = s.uploadManager.SaveFile(ctx, internalPath, img.Data); err != nil {
		observability.AcknowledgeError(err, logger, span, "saving provided avatar")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	user.AvatarSrc = stringPointer(internalPath)

	if err = s.userDataManager.UpdateUser(ctx, user, nil); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating user info")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}
}

// ArchiveHandler is a handler for archiving a user.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out who this is for.
	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	// do the deed.
	err := s.userDataManager.ArchiveUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// inform the relatives.
	s.userCounter.Decrement(ctx)

	// we're all good.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// figure out who this is for.
	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	x, err := s.userDataManager.GetAuditLogEntriesForUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching audit log entries for user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}
