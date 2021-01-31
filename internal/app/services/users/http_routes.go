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

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	passwordvalidator "github.com/lane-c-wagner/go-password-validator"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp/totp"
)

const (
	// UserIDURIParamKey is used to refer to user IDs in router params.
	UserIDURIParamKey = "userID"

	totpIssuer        = "todoService"
	base64ImagePrefix = "data:image/jpeg;base64,"

	minimumPasswordEntropy = 75
)

// validateCredentialChangeRequest takes a user's credentials and determines
// if they match what is on record.
func (s *service) validateCredentialChangeRequest(
	ctx context.Context,
	userID uint64,
	password,
	totpToken string,
) (user *types.User, httpStatus int) {
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
	if valid, validationErr := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		password,
		user.TwoFactorSecret,
		totpToken,
		user.Salt,
	); validationErr != nil {
		logger.Error(err, "error encountered validating credentials")
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
		logger.Error(err, "error fetching users for UsernameSearchHandler route")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response.
	s.encoderDecoder.EncodeResponse(ctx, res, users)
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
		logger.Error(err, "error fetching users for ListHandler route")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response.
	s.encoderDecoder.EncodeResponse(ctx, res, users)
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

	// fetch parsed input from request context.
	userInput, ok := ctx.Value(userCreationMiddlewareCtxKey).(*types.NewUserCreationInput)
	if !ok {
		logger.Info("valid input not attached to UsersService CreateHandler request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// NOTE: I feel comfortable letting username be in the logger, since
	// the logging statements below are only in the event of errors. If
	// and when that changes, this can/should be removed.
	logger = logger.WithValue(keys.UsernameKey, userInput.Username)
	tracing.AttachUsernameToSpan(span, userInput.Username)

	// ensure the password isn't garbage-tier
	if err := passwordvalidator.Validate(userInput.Password, minimumPasswordEntropy); err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "password too weak!", http.StatusBadRequest)
		return
	}

	// hash the password.
	hp, err := s.authenticator.HashPassword(ctx, userInput.Password)
	if err != nil {
		logger.Error(err, "valid input not attached to request")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	input := types.UserDataStoreCreationInput{
		Username:        userInput.Username,
		HashedPassword:  hp,
		TwoFactorSecret: "",
		Salt:            []byte{},
	}

	// generate a two factor secret.
	if input.TwoFactorSecret, err = s.secretGenerator.GenerateTwoFactorSecret(); err != nil {
		logger.Error(err, "error generating TOTP secret")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// generate a salt.
	if input.Salt, err = s.secretGenerator.GenerateSalt(); err != nil {
		logger.Error(err, "error generating salt")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// create the user.
	user, userCreationErr := s.userDataManager.CreateUser(ctx, input)
	if userCreationErr != nil {
		logger.Error(userCreationErr, "error creating user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	if _, accountCreationErr := s.accountDataManager.CreateAccount(ctx, &types.AccountCreationInput{
		Name:          user.Username,
		BelongsToUser: user.ID,
	}); accountCreationErr != nil {
		logger.Error(accountCreationErr, "error creating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// UserCreationResponse is a struct we can use to notify the user of
	// their two factor secret, but ideally just this once and then never again.
	ucr := &types.UserCreationResponse{
		ID:              user.ID,
		Username:        user.Username,
		CreatedOn:       user.CreatedOn,
		TwoFactorQRCode: s.buildQRCode(ctx, user.Username, user.TwoFactorSecret),
	}

	// notify the relevant parties.
	tracing.AttachUserIDToSpan(span, user.ID)
	s.userCounter.Increment(ctx)
	s.auditLog.LogUserCreationEvent(ctx, user)

	// encode and peace.
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, ucr, http.StatusCreated)
}

// buildQRCode builds a QR code for a given username and secret.
func (s *service) buildQRCode(ctx context.Context, username, twoFactorSecret string) string {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	// "otpauth://totp/{{ .Issuer }}:{{ .EnsureUsername }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
	otpString := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		totpIssuer,
		username,
		twoFactorSecret,
		totpIssuer,
	)

	bmp, err := qrcode.NewQRCodeWriter().EncodeWithoutHint(otpString, gozxing.BarcodeFormat_QR_CODE, 128, 128)
	if err != nil {
		s.logger.Error(err, "trying to encode secret to qr code")
		return ""
	}

	// encode the QR code to PNG.
	var b bytes.Buffer
	if err = png.Encode(&b, bmp); err != nil {
		s.logger.Error(err, "trying to encode secret to qr code")
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

	si, err := s.sessionInfoFetcher(req)
	if err != nil {
		logger.Error(err, "session info missing from request context")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	// figure out who this is all for.
	userID := si.UserID
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	// fetch user data.
	x, err := s.userDataManager.GetUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no such user")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching user from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
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
		logger.Error(err, "error fetching user from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// TOTPSecretVerificationHandler accepts a TOTP token as input and returns 200 if the TOTP token
// is validated by the user's TOTP secret.
func (s *service) TOTPSecretVerificationHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed input.
	input, ok := req.Context().Value(totpSecretVerificationMiddlewareCtxKey).(*types.TOTPSecretVerificationInput)
	if !ok || input == nil {
		logger.Debug("no input found on TOTP secret refresh request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	user, err := s.userDataManager.GetUserWithUnverifiedTwoFactorSecret(ctx, input.UserID)
	if err != nil {
		logger.Error(err, "fetching user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachUserIDToSpan(span, user.ID)
	tracing.AttachUsernameToSpan(span, user.Username)

	if user.TwoFactorSecretVerifiedOn != nil {
		// I suppose if this happens too many times, we'll want to keep track of that
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "TOTP secret already verified", http.StatusAlreadyReported)
		return
	}

	statusCode := http.StatusBadRequest

	if totp.Validate(input.TOTPToken, user.TwoFactorSecret) {
		statusCode = http.StatusAccepted

		if updateUserErr := s.userDataManager.VerifyUserTwoFactorSecret(ctx, user.ID); updateUserErr != nil {
			logger.Error(updateUserErr, "updating user to indicate their 2FA secret is validated")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			return
		}

		s.auditLog.LogUserVerifyTwoFactorSecretEvent(ctx, user.ID)
	}

	res.WriteHeader(statusCode)
}

// NewTOTPSecretHandler fetches a user, and issues them a new TOTP secret, after validating
// that information received from TOTPSecretRefreshInputContextMiddleware is valid.
func (s *service) NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed input.
	input, ok := req.Context().Value(totpSecretRefreshMiddlewareCtxKey).(*types.TOTPSecretRefreshInput)
	if !ok {
		logger.Debug("no input found on TOTP secret refresh request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// also check for the user's ID.
	si, ok := ctx.Value(types.SessionInfoKey).(*types.SessionInfo)
	if !ok || si == nil {
		logger.Debug("no user ID attached to TOTP secret refresh request")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	// make sure this is all on the up-and-up
	user, httpStatus := s.validateCredentialChangeRequest(
		ctx,
		si.UserID,
		input.CurrentPassword,
		input.TOTPToken,
	)

	// if the above function returns something other than 200, it means some error occurred.
	if httpStatus != http.StatusOK {
		res.WriteHeader(httpStatus)
		return
	}

	// document who this is for.
	tracing.AttachUserIDToSpan(span, si.UserID)
	tracing.AttachUsernameToSpan(span, user.Username)
	logger = logger.WithValue(keys.UserIDKey, user.ID)

	// set the two factor secret.
	tfs, err := s.secretGenerator.GenerateTwoFactorSecret()
	if err != nil {
		logger.Error(err, "error encountered generating random TOTP string")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	user.TwoFactorSecret = tfs
	user.TwoFactorSecretVerifiedOn = nil

	// update the user in the database.
	if err := s.userDataManager.UpdateUser(ctx, user); err != nil {
		logger.Error(err, "error encountered updating TOTP token")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// let the requester know we're all good.
	result := &types.TOTPSecretRefreshResponse{
		TwoFactorSecret: user.TwoFactorSecret,
		TwoFactorQRCode: s.buildQRCode(ctx, user.Username, user.TwoFactorSecret),
	}

	s.auditLog.LogUserUpdateTwoFactorSecretEvent(ctx, user.ID)

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, result, http.StatusAccepted)
}

// UpdatePasswordHandler updates a user's password, after validating that information received
// from PasswordUpdateInputContextMiddleware is valid.
func (s *service) UpdatePasswordHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed value.
	input, ok := ctx.Value(passwordChangeMiddlewareCtxKey).(*types.PasswordUpdateInput)
	if !ok {
		logger.Debug("no input found on UpdatePasswordHandler request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// check request context for user ID.
	si, ok := ctx.Value(types.SessionInfoKey).(*types.SessionInfo)
	if !ok || si == nil {
		logger.Debug("no user ID attached to UpdatePasswordHandler request")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	// determine relevant user ID.
	tracing.AttachUserIDToSpan(span, si.UserID)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// make sure everything's on the up-and-up
	user, httpStatus := s.validateCredentialChangeRequest(
		ctx,
		si.UserID,
		input.CurrentPassword,
		input.TOTPToken,
	)

	// if the above function returns something other than 200, it means some error occurred.
	if httpStatus != http.StatusOK {
		res.WriteHeader(httpStatus)
		return
	}

	tracing.AttachUsernameToSpan(span, user.Username)

	// ensure the password isn't garbage-tier
	if err := passwordvalidator.Validate(input.NewPassword, minimumPasswordEntropy); err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "new password is too weak!", http.StatusBadRequest)
		return
	}

	// hash the new password.
	newPasswordHash, err := s.authenticator.HashPassword(ctx, input.NewPassword)
	if err != nil {
		logger.Error(err, "error hashing password")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the user.
	if err = s.userDataManager.UpdateUserPassword(ctx, user.ID, newPasswordHash); err != nil {
		logger.Error(err, "error encountered updating user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// we're all good, log the user out
	cookie, cookieRetrievalErr := req.Cookie(s.authSettings.Cookies.Name)
	if cookieRetrievalErr != nil {
		// this should never occur in production
		logger.Error(cookieRetrievalErr, "retrieving cookie to invalidate upon request")
	} else {
		cookie.MaxAge = -1
		http.SetCookie(res, cookie)
	}

	s.auditLog.LogUserUpdatePasswordEvent(ctx, user.ID)

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections#Temporary_redirections
	http.Redirect(res, req, "/auth/login", http.StatusSeeOther)
}

func (s *service) determineAvatarWebPath(storageProviderPath string) *string {
	x := storageProviderPath
	return &x
}

// AvatarUploadHandler updates a user's avatar.
func (s *service) AvatarUploadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for user ID.
	si, ok := ctx.Value(types.SessionInfoKey).(*types.SessionInfo)
	if !ok || si == nil {
		logger.Debug("no user ID attached to UpdateAvatarHandler request")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	user, userFetchErr := s.userDataManager.GetUser(ctx, si.UserID)
	if userFetchErr != nil {
		logger.Error(userFetchErr, "fetching associated user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	img, imageProcessErr := s.imageUploadProcessor.Process(ctx, req, "avatar")
	if imageProcessErr != nil || img == nil {
		logger.Error(imageProcessErr, "processing provided avatar upload file")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	internalPath := fmt.Sprintf("avatar_%d", si.UserID)

	if saveErr := s.uploadManager.SaveFile(ctx, internalPath, img.Data); saveErr != nil {
		logger.WithValue("file_size", len(img.Data)).Error(saveErr, "saving provided avatar")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	user.AvatarSrc = s.determineAvatarWebPath(internalPath)

	if userUpdateErr := s.userDataManager.UpdateUser(ctx, user); userUpdateErr != nil {
		logger.WithValue("file_size", len(img.Data)).Error(userUpdateErr, "updating user info")
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
	if err := s.userDataManager.ArchiveUser(ctx, userID); err != nil {
		logger.Error(err, "deleting user from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// inform the relatives.
	s.userCounter.Decrement(ctx)
	s.auditLog.LogUserArchiveEvent(ctx, userID)

	// we're all good.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// figure out who this is for.
	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	x, err := s.auditLog.GetAuditLogEntriesForUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}
