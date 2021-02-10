package querier

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountSubscriptionPlanDataManager = (*Client)(nil)
)

// scanPlan takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountSubscriptionPlan struct.
func (c *Client) scanAccountSubscriptionPlan(scan database.Scanner, includeCounts bool) (plan *types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	plan = &types.AccountSubscriptionPlan{}

	var rawPeriod string

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

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	p, parseErr := time.ParseDuration(rawPeriod)
	if parseErr != nil {
		return nil, 0, 0, parseErr
	}

	plan.Period = p

	return plan, filteredCount, totalCount, nil
}

// scanAccountSubscriptionPlans takes some database rows and turns them into a slice of account subscription plans.
func (c *Client) scanAccountSubscriptionPlans(rows database.ResultIterator, includeCounts bool) (plans []*types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := c.scanAccountSubscriptionPlan(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
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

	if handleErr := c.handleRows(rows); handleErr != nil {
		return nil, 0, 0, handleErr
	}

	return plans, filteredCount, totalCount, nil
}

// GetAccountSubscriptionPlan fetches a plan from the database.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID uint64) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID).Debug("GetAccountSubscriptionPlan called")
	tracing.AttachPlanIDToSpan(span, accountSubscriptionPlanID)

	query, args := c.sqlQueryBuilder.BuildGetAccountSubscriptionPlanQuery(accountSubscriptionPlanID)
	row := c.db.QueryRowContext(ctx, query, args...)

	plan, _, _, err := c.scanAccountSubscriptionPlan(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning account subscription plan: %w", err)
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

	x = &types.AccountSubscriptionPlanList{}

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	c.logger.Debug("GetAccountSubscriptionPlans called")

	query, args := c.sqlQueryBuilder.BuildGetAccountSubscriptionPlansQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing account subscription plan list retrieval query: %w", err)
	}

	if x.AccountSubscriptionPlans, x.FilteredCount, x.TotalCount, err = c.scanAccountSubscriptionPlans(rows, true); err != nil {
		return nil, fmt.Errorf("scanning account subscription account subscription plans: %w", err)
	}

	return x, nil
}

// CreateAccountSubscriptionPlan creates a plan in the database.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildCreateAccountSubscriptionPlanQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the account subscription plan.
	id, err := c.performWriteQuery(ctx, tx, false, "account subscription plan creation", query, args)
	if err != nil {
		return nil, err
	}

	x := &types.AccountSubscriptionPlan{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Period:      input.Period,
		CreatedOn:   c.currentTime(),
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanCreationEventEntry(x))

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return x, nil
}

// UpdateAccountSubscriptionPlan updates a particular plan. Note that UpdatePlan expects the provided input to have a valid ID.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, updated *types.AccountSubscriptionPlan, changedBy uint64, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, updated.ID).Debug("UpdateAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildUpdateAccountSubscriptionPlanQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, c.db, "account subscription plan update", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error updating account subscription plan: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanUpdateEventEntry(changedBy, updated.ID, changes))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// ArchiveAccountSubscriptionPlan archives a plan from the database by its ID.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID, archivedBy uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, accountSubscriptionPlanID)

	c.logger.WithValues(map[string]interface{}{
		keys.AccountSubscriptionPlanIDKey: accountSubscriptionPlanID,
		keys.UserIDKey:                    archivedBy,
	}).Debug("ArchiveAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildArchiveAccountSubscriptionPlanQuery(accountSubscriptionPlanID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, c.db, "account subscription plan archive", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error updating account subscription plan: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountSubscriptionPlanArchiveEventEntry(archivedBy, accountSubscriptionPlanID))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetAuditLogEntriesForAccountSubscriptionPlan fetches a list of audit log entries from the database that relate to a given plan.
func (c *Client) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, accountSubscriptionPlanID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID).Debug("GetAuditLogEntriesForAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(accountSubscriptionPlanID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning audit log entries: %w", err)
	}

	return auditLogEntries, nil
}
