package querier

import (
	"context"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountSubscriptionPlanDataManager = (*SQLQuerier)(nil)
)

// scanPlan takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountSubscriptionPlan struct.
func (q *SQLQuerier) scanAccountSubscriptionPlan(ctx context.Context, scan database.Scanner, includeCounts bool) (plan *types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	var rawPeriod string
	plan = &types.AccountSubscriptionPlan{}

	targetVars := []interface{}{
		&plan.ID,
		&plan.ExternalID,
		&plan.Name,
		&plan.Description,
		&plan.Price,
		&rawPeriod,
		&plan.CreatedOn,
		&plan.LastUpdatedOn,
		&plan.ArchivedOn,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "scanning account")
	}

	p, err := time.ParseDuration(rawPeriod)
	if err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "parsing stored account subscription plan period duration")
	}

	plan.Period = p

	return plan, filteredCount, totalCount, nil
}

// scanAccountSubscriptionPlans takes some database rows and turns them into a slice of account subscription plans.
func (q *SQLQuerier) scanAccountSubscriptionPlans(ctx context.Context, rows database.ResultIterator, includeCounts bool) (plans []*types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		x, fc, tc, scanErr := q.scanAccountSubscriptionPlan(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, observability.PrepareError(scanErr, logger, span, "scanning account subscription plan")
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		plans = append(plans, x)
	}

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return plans, filteredCount, totalCount, nil
}

// GetAccountSubscriptionPlan fetches a plan from the database.
func (q *SQLQuerier) GetAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID uint64) (*types.AccountSubscriptionPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountSubscriptionPlanID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)

	query, args := q.sqlQueryBuilder.BuildGetAccountSubscriptionPlanQuery(ctx, accountSubscriptionPlanID)
	row := q.db.QueryRowContext(ctx, query, args...)

	plan, _, _, err := q.scanAccountSubscriptionPlan(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning account subscription plan")
	}

	return plan, nil
}

// GetAllAccountSubscriptionPlansCount fetches the count of account subscription plans from the database that meet a particular filter.
func (q *SQLQuerier) GetAllAccountSubscriptionPlansCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllAccountSubscriptionPlansCountQuery(ctx), "fetching count of account subscription plans")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of account subscription plans")
	}

	return count, nil
}

// GetAccountSubscriptionPlans fetches a list of account subscription plans from the database that meet a particular filter.
func (q *SQLQuerier) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (x *types.AccountSubscriptionPlanList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(q.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.AccountSubscriptionPlanList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetAccountSubscriptionPlansQuery(ctx, filter)

	rows, err := q.performReadQuery(ctx, "account subscription plan", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying for account subscription plans")
	}

	if x.AccountSubscriptionPlans, x.FilteredCount, x.TotalCount, err = q.scanAccountSubscriptionPlans(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning account subscription plans")
	}

	return x, nil
}

// CreateAccountSubscriptionPlan creates a plan in the database.
func (q *SQLQuerier) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	query, args := q.sqlQueryBuilder.BuildCreateAccountSubscriptionPlanQuery(ctx, input)
	logger := q.logger.WithValue(keys.NameKey, input.Name)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the account subscription plan.
	id, err := q.performWriteQuery(ctx, tx, false, "account subscription plan creation", query, args)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account subscription plan")
	}

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, id)

	accountSubscriptionPlan := &types.AccountSubscriptionPlan{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Period:      input.Period,
		CreatedOn:   q.currentTime(),
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanCreationEventEntry(accountSubscriptionPlan)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing account subscription plan creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	return accountSubscriptionPlan, nil
}

// UpdateAccountSubscriptionPlan updates a particular plan. Note that UpdatePlan expects the provided input to have a valid ID.
func (q *SQLQuerier) UpdateAccountSubscriptionPlan(ctx context.Context, updated *types.AccountSubscriptionPlan, changedBy uint64, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.AccountSubscriptionPlanIDKey, updated.ID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, updated.ID)
	tracing.AttachRequestingUserIDToSpan(span, changedBy)
	tracing.AttachChangeSummarySpan(span, "account", changes)

	query, args := q.sqlQueryBuilder.BuildUpdateAccountSubscriptionPlanQuery(ctx, updated)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "account subscription plan update", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating account subscription plan")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanUpdateEventEntry(changedBy, updated.ID, changes)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account subscription plan update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// ArchiveAccountSubscriptionPlan archives a plan from the database by its ID.
func (q *SQLQuerier) ArchiveAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID, archivedBy uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountSubscriptionPlanID == 0 {
		return ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)
	tracing.AttachRequestingUserIDToSpan(span, archivedBy)

	query, args := q.sqlQueryBuilder.BuildArchiveAccountSubscriptionPlanQuery(ctx, accountSubscriptionPlanID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "account subscription plan archive", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating account subscription plan")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanArchiveEventEntry(archivedBy, accountSubscriptionPlanID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account subscription plan archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForAccountSubscriptionPlan fetches a list of audit log entries from the database that relate to a given plan.
func (q *SQLQuerier) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountSubscriptionPlanID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)

	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(ctx, accountSubscriptionPlanID)

	rows, err := q.performReadQuery(ctx, "audit log entries for account subscription plan", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entries")
	}

	return auditLogEntries, nil
}
