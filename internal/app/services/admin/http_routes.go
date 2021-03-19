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

	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	if !reqCtx.User.ServiceAdminPermissions.IsServiceAdmin() {
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	}

	logger = logger.WithValue("ban_giver", reqCtx.User.ID)

	var allowed bool

	switch input.NewReputation {
	case types.BannedAccountStatus:
		allowed = reqCtx.User.ServiceAdminPermissions.CanBanUsers()
	case types.TerminatedAccountStatus:
		allowed = reqCtx.User.ServiceAdminPermissions.CanTerminateAccounts()
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	if !allowed {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	logger = logger.WithValue("status_change_recipient", input.TargetUserID)

	if err := s.userDB.UpdateUserAccountStatus(ctx, input.TargetUserID, *input); err != nil {
		logger.Error(err, "changing user status")

		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		} else {
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		}

		return
	}

	switch input.NewReputation {
	case types.BannedAccountStatus:
		s.auditLog.LogUserBanEvent(ctx, reqCtx.User.ID, input.TargetUserID, input.Reason)
	case types.TerminatedAccountStatus:
		s.auditLog.LogAccountTerminationEvent(ctx, reqCtx.User.ID, input.TargetUserID, input.Reason)
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, nil, http.StatusAccepted)
}
