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
	input, ok := ctx.Value(accountStatusUpdateMiddlewareCtxKey).(*types.UserReputationUpdateInput)
	if !ok || input == nil {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	logger = logger.WithValue("new_status", input.NewReputation)

	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		logger.Error(sessionInfoRetrievalErr, "error fetching sessionInfo")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	if !si.ServiceAdminPermissions.IsServiceAdmin() {
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue("ban_giver", si.UserID)

	var allowed bool

	switch input.NewReputation {
	case types.BannedAccountStatus:
		allowed = si.ServiceAdminPermissions.CanBanUsers()
	case types.TerminatedAccountStatus:
		allowed = si.ServiceAdminPermissions.CanTerminateAccounts()
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	if !allowed {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	logger = logger.WithValue("ban_recipient", input.TargetAccountID)

	if err := s.userDB.UpdateUserAccountStatus(ctx, input.TargetAccountID, *input); err != nil {
		logger.Error(err, "error banning user")

		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		} else {
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		}

		return
	}

	switch input.NewReputation {
	case types.BannedAccountStatus:
		s.auditLog.LogUserBanEvent(ctx, si.UserID, input.TargetAccountID, input.Reason)
	case types.TerminatedAccountStatus:
		s.auditLog.LogAccountTerminationEvent(ctx, si.UserID, input.TargetAccountID, input.Reason)
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, nil, http.StatusAccepted)
}
