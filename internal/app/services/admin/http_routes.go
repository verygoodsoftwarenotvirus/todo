package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// UserIDURIParamKey is used to refer to user IDs in router params.
	UserIDURIParamKey = "userID"
)

// UserAccountStatusChangeHandler changes a user's status.
func (s *service) UserAccountStatusChangeHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(accountStatusUpdateMiddlewareCtxKey).(*types.AccountStatusUpdateInput)
	if !ok || input == nil {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeNoInputResponse(res)
		return
	}

	logger = logger.WithValue("new_status", input.NewStatus)

	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	if !si.UserIsAdmin {
		s.encoderDecoder.EncodeUnauthorizedResponse(res)
		return
	}

	logger = logger.WithValue("ban_giver", si.UserID)

	var allowed bool

	switch input.NewStatus {
	case types.BannedAccountStatus:
		allowed = si.AdminPermissions.CanBanUsers()
	case types.TerminatedAccountStatus:
		allowed = si.AdminPermissions.CanTerminateAccounts()
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	if !allowed {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(res)
		return
	}

	logger = logger.WithValue("ban_recipient", input.TargetAccountID)

	if err := s.userDB.UpdateUserAccountStatus(ctx, input.TargetAccountID, *input); err != nil {
		logger.Error(err, "error banning user")

		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(res)
		} else {
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		}

		return
	}

	switch input.NewStatus {
	case types.BannedAccountStatus:
		s.auditLog.LogUserBanEvent(ctx, si.UserID, input.TargetAccountID, input.Reason)
	case types.TerminatedAccountStatus:
		s.auditLog.LogAccountTerminationEvent(ctx, si.UserID, input.TargetAccountID, input.Reason)
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	s.encoderDecoder.EncodeResponseWithStatus(res, nil, http.StatusAccepted)
}
