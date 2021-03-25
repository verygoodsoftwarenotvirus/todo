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
	_ types.AccountSubscriptionPlanDataManager = (*Client)(nil)
)

// scanPlan takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountSubscriptionPlan struct.
func (c *Client) scanAccountSubscriptionPlan(ctx context.Context, scan database.Scanner, includeCounts bool) (plan *types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

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
func (c *Client) scanAccountSubscriptionPlans(ctx context.Context, rows database.ResultIterator, includeCounts bool) (plans []*types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		x, fc, tc, scanErr := c.scanAccountSubscriptionPlan(ctx, rows, includeCounts)
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

	if err = c.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return plans, filteredCount, totalCount, nil
}

// GetAccountSubscriptionPlan fetches a plan from the database.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID uint64) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)

	logger.Debug("GetAccountSubscriptionPlan called")
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)

	query, args := c.sqlQueryBuilder.BuildGetAccountSubscriptionPlanQuery(accountSubscriptionPlanID)
	row := c.db.QueryRowContext(ctx, query, args...)

	plan, _, _, err := c.scanAccountSubscriptionPlan(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning account subscription plan")
	}

	return plan, nil
}

// GetAllAccountSubscriptionPlansCount fetches the count of account subscription plans from the database that meet a particular filter.
func (c *Client) GetAllAccountSubscriptionPlansCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountSubscriptionPlansCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllAccountSubscriptionPlansCountQuery())
}

// GetAccountSubscriptionPlans fetches a list of account subscription plans from the database that meet a particular filter.
func (c *Client) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (x *types.AccountSubscriptionPlanList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(c.logger)

	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.AccountSubscriptionPlanList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAccountSubscriptionPlansQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying for account subscription plans")
	}

	if x.AccountSubscriptionPlans, x.FilteredCount, x.TotalCount, err = c.scanAccountSubscriptionPlans(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning account subscription plans")
	}

	return x, nil
}

// CreateAccountSubscriptionPlan creates a plan in the database.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	query, args := c.sqlQueryBuilder.BuildCreateAccountSubscriptionPlanQuery(input)
	logger := c.logger.WithValue(keys.NameKey, input.Name)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the account subscription plan.
	id, err := c.performWriteQuery(ctx, tx, false, "account subscription plan creation", query, args)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account subscription plan")
	}

	x := &types.AccountSubscriptionPlan{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Period:      input.Period,
		CreatedOn:   c.currentTime(),
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanCreationEventEntry(x)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing account subscription plan creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	return x, nil
}

// UpdateAccountSubscriptionPlan updates a particular plan. Note that UpdatePlan expects the provided input to have a valid ID.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, updated *types.AccountSubscriptionPlan, changedBy uint64, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, updated.ID)
	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, updated.ID)

	query, args := c.sqlQueryBuilder.BuildUpdateAccountSubscriptionPlanQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "account subscription plan update", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating account subscription plan")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanUpdateEventEntry(changedBy, updated.ID, changes)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account subscription plan update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// ArchiveAccountSubscriptionPlan archives a plan from the database by its ID.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID, archivedBy uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.AccountSubscriptionPlanIDKey: accountSubscriptionPlanID,
		keys.UserIDKey:                    archivedBy,
	})

	query, args := c.sqlQueryBuilder.BuildArchiveAccountSubscriptionPlanQuery(accountSubscriptionPlanID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "account subscription plan archive", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating account subscription plan")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanArchiveEventEntry(archivedBy, accountSubscriptionPlanID)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account subscription plan archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForAccountSubscriptionPlan fetches a list of audit log entries from the database that relate to a given plan.
func (c *Client) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(accountSubscriptionPlanID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entries")
	}

	return auditLogEntries, nil
}
