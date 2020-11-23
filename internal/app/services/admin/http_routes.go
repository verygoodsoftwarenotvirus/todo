package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
)

const (
	// UserIDURIParamKey is used to refer to user IDs in router params.
	UserIDURIParamKey = "userID"
)

// BanHandler bans a user.
func (s *Service) BanHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	logger = logger.WithValue("ban_giver", si.UserID)

	if !si.AdminPermissions.CanBanUsers() || !si.UserIsAdmin {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(res)
		return
	}

	banRecipient := s.userIDFetcher(req)
	logger = logger.WithValue("ban_recipient", banRecipient)

	if err := s.userDB.BanUser(ctx, banRecipient); err != nil {
		logger.Error(err, "error banning user")

		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(res)
		} else {
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		}
		return
	}

	s.auditLog.LogUserBanEvent(ctx, si.UserID, banRecipient)
	s.encoderDecoder.EncodeResponseWithStatus(res, nil, http.StatusAccepted)
}

// AccountTerminationHandler terminates an account.
func (s *Service) AccountTerminationHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	logger = logger.WithValue("terminator", si.UserID)

	if !si.AdminPermissions.CanBanUsers() || !si.UserIsAdmin {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(res)
		return
	}

	banRecipient := s.userIDFetcher(req)
	logger = logger.WithValue("terminee", banRecipient)

	if err := s.userDB.BanUser(ctx, banRecipient); err != nil {
		logger.Error(err, "error terminating account")

		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(res)
		} else {
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		}
		return
	}

	s.auditLog.LogUserBanEvent(ctx, si.UserID, banRecipient)
	s.encoderDecoder.EncodeResponseWithStatus(res, nil, http.StatusAccepted)
}
