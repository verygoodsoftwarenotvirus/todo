package querier

import (
	"context"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ types.AdminUserDataManager = (*SQLQuerier)(nil)

// UpdateUserReputation updates a user's account status.
func (q *SQLQuerier) UpdateUserReputation(ctx context.Context, userID uint64, input *types.UserReputationUpdateInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	query, args := q.sqlQueryBuilder.BuildSetUserStatusQuery(ctx, input)

	if err := q.performWriteQueryIgnoringReturn(ctx, q.db, "user status update query", query, args); err != nil {
		return observability.PrepareError(err, logger, span, "user status update")
	}

	logger.Info("user reputation updated")

	return nil
}

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (q *SQLQuerier) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64, reason string) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, banRecipient)

	q.createAuditLogEntry(ctx, q.db, audit.BuildUserBanEventEntry(banGiver, banRecipient, reason))
}

// LogAccountTerminationEvent saves a UserBannedEvent in the audit log table.
func (q *SQLQuerier) LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, terminee)

	q.createAuditLogEntry(ctx, q.db, audit.BuildAccountTerminationEventEntry(terminator, terminee, reason))
}

// LogCycleCookieSecretEvent implements our AuditLogEntryDataManager interface.
func (q *SQLQuerier) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	q.createAuditLogEntry(ctx, q.db, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent implements our AuditLogEntryDataManager interface.
func (q *SQLQuerier) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	q.createAuditLogEntry(ctx, q.db, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogBannedUserLoginAttemptEvent implements our AuditLogEntryDataManager interface.
func (q *SQLQuerier) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	q.createAuditLogEntry(ctx, q.db, audit.BuildBannedUserLoginAttemptEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent implements our AuditLogEntryDataManager interface.
func (q *SQLQuerier) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	q.createAuditLogEntry(ctx, q.db, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our AuditLogEntryDataManager interface.
func (q *SQLQuerier) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	q.createAuditLogEntry(ctx, q.db, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent implements our AuditLogEntryDataManager interface.
func (q *SQLQuerier) LogLogoutEvent(ctx context.Context, userID uint64) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	q.createAuditLogEntry(ctx, q.db, audit.BuildLogoutEventEntry(userID))
}
