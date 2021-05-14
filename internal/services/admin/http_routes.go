package admin

import (
	"database/sql"
	"errors"
	"net/http"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
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
	tracing.AttachRequestToSpan(span, req)

	// check session context data for parsed input struct.
	input, ok := ctx.Value(accountStatusUpdateMiddlewareCtxKey).(*types.UserReputationUpdateInput)
	if !ok || input == nil {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	logger = logger.WithValue("new_status", input.NewReputation)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

	if !sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin() {
		// this should never happen in production
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "inadequate permissions for route", http.StatusForbidden)
		return
	}

	requester := sessionCtxData.Requester.ID
	logger = logger.WithValue("ban_giver", requester)

	var allowed bool

	switch input.NewReputation {
	case types.BannedUserReputation:
		allowed = sessionCtxData.Requester.ServiceAdminPermission.CanBanUsers()
	case types.TerminatedUserReputation:
		allowed = sessionCtxData.Requester.ServiceAdminPermission.CanTerminateAccounts()
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
		allowed = true
	}

	if !allowed {
		logger.Info("ban attempt made by admin without appropriate permissions")
		s.encoderDecoder.EncodeInvalidPermissionsResponse(ctx, res)
		return
	}

	logger = logger.WithValue("status_change_recipient", input.TargetUserID)

	if err = s.userDB.UpdateUserReputation(ctx, input.TargetUserID, input); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		} else {
			observability.AcknowledgeError(err, logger, span, "retrieving session context data")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		}

		return
	}

	switch input.NewReputation {
	case types.BannedUserReputation:
		s.auditLog.LogUserBanEvent(ctx, requester, input.TargetUserID, input.Reason)
	case types.TerminatedUserReputation:
		s.auditLog.LogAccountTerminationEvent(ctx, requester, input.TargetUserID, input.Reason)
	case types.GoodStandingAccountStatus, types.UnverifiedAccountStatus:
		// the appropriate audit log entry is already written, the above are supplementary
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, nil, http.StatusAccepted)
}
