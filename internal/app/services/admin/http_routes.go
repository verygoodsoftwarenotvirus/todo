package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
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

	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)

	if !reqCtx.Requester.ServiceAdminPermissions.IsServiceAdmin() {
		// this should never happen in production
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "inadequate permissions for route", http.StatusForbidden)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue("ban_giver", requester)

	var allowed bool

	switch input.NewReputation {
	case types.BannedAccountStatus:
		allowed = reqCtx.Requester.ServiceAdminPermissions.CanBanUsers()
	case types.TerminatedAccountStatus:
		allowed = reqCtx.Requester.ServiceAdminPermissions.CanTerminateAccounts()
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	if !allowed {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	logger = logger.WithValue("status_change_recipient", input.TargetUserID)

	if err = s.userDB.UpdateUserReputation(ctx, input.TargetUserID, *input); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		} else {
			observability.AcknowledgeError(err, logger, span, "retrieving request context")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		}

		return
	}

	switch input.NewReputation {
	case types.BannedAccountStatus:
		s.auditLog.LogUserBanEvent(ctx, requester, input.TargetUserID, input.Reason)
	case types.TerminatedAccountStatus:
		s.auditLog.LogAccountTerminationEvent(ctx, requester, input.TargetUserID, input.Reason)
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, nil, http.StatusAccepted)
}
