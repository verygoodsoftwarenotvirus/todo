package superclient

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountSubscriptionPlanDataManager  = (*Client)(nil)
	_ types.AccountSubscriptionPlanAuditManager = (*Client)(nil)
)

// scanPlan takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountSubscriptionPlan struct.
func (c *Client) scanAccountSubscriptionPlan(scan database.Scanner, includeCounts bool) (plan *types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	plan = &types.AccountSubscriptionPlan{}

	var rawPeriod string

	targetVars := []interface{}{
		&plan.ID,
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

// scanPlans takes some database rows and turns them into a slice of plans.
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

	if rowErr := rows.Err(); rowErr != nil {
		return nil, 0, 0, rowErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing database rows")
		return nil, 0, 0, closeErr
	}

	return plans, filteredCount, totalCount, nil
}

// GetAccountSubscriptionPlan fetches an plan from the database.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, planID)

	c.logger.WithValue("plan_id", planID).Debug("GetAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildGetAccountSubscriptionPlanQuery(planID)
	row := c.db.QueryRowContext(ctx, query, args...)

	plan, _, _, err := c.scanAccountSubscriptionPlan(row, false)

	return plan, err
}

// GetAllAccountSubscriptionPlansCount fetches the count of plans from the database that meet a particular filter.
func (c *Client) GetAllAccountSubscriptionPlansCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountSubscriptionPlansCount called")

	err = c.db.QueryRowContext(ctx, c.sqlQueryBuilder.BuildGetAllAccountSubscriptionPlansCountQuery()).Scan(&count)
	return count, err
}

// GetAccountSubscriptionPlans fetches a list of plans from the database that meet a particular filter.
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
		return nil, fmt.Errorf("querying database for plans: %w", err)
	}

	x.AccountSubscriptionPlans, x.FilteredCount, x.TotalCount, err = c.scanAccountSubscriptionPlans(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return x, nil
}

// CreateAccountSubscriptionPlan creates an plan in the database.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateAccountSubscriptionPlan called")

	x := &types.AccountSubscriptionPlan{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Period:      input.Period,
	}

	query, args := c.sqlQueryBuilder.BuildCreateAccountSubscriptionPlanQuery(x)

	// create the plan.
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	x.CreatedOn = c.timeTeller.Now()
	x.ID = c.getIDFromResult(res)

	return x, nil
}

// UpdateAccountSubscriptionPlan updates a particular plan. Note that UpdatePlan expects the provided input to have a valid ID.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, updated *types.AccountSubscriptionPlan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, updated.ID).Debug("UpdateAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildUpdateAccountSubscriptionPlanQuery(updated)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// ArchiveAccountSubscriptionPlan archives an plan from the database by its ID.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, planID)

	c.logger.WithValues(map[string]interface{}{
		"plan_id": planID,
	}).Debug("ArchiveAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildArchiveAccountSubscriptionPlanQuery(planID)

	res, err := c.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogAccountSubscriptionPlanCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountSubscriptionPlanCreationEvent(ctx context.Context, plan *types.AccountSubscriptionPlan) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("LogAccountSubscriptionPlanCreationEvent called")

	c.createAuditLogEntry(ctx, audit.BuildAccountSubscriptionPlanCreationEventEntry(plan))
}

// AccountSubscriptionLogPlanUpdateEvent implements our AuditLogEntryDataManager interface.
func (c *Client) AccountSubscriptionLogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("AccountSubscriptionLogPlanUpdateEvent called")

	c.createAuditLogEntry(ctx, audit.BuildAccountSubscriptionPlanUpdateEventEntry(userID, planID, changes))
}

// AccountSubscriptionLogPlanArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) AccountSubscriptionLogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("AccountSubscriptionLogPlanArchiveEvent called")

	c.createAuditLogEntry(ctx, audit.BuildAccountSubscriptionPlanArchiveEventEntry(userID, planID))
}

// GetAuditLogEntriesForAccountSubscriptionPlan fetches a list of audit log entries from the database that relate to a given plan.
func (c *Client) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID).Debug("GetAuditLogEntriesForAccountSubscriptionPlan called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(planID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}
