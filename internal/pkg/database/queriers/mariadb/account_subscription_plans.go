package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AccountSubscriptionPlanDataManager = (*MariaDB)(nil)

// scanPlan takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountSubscriptionPlan struct.
func (q *MariaDB) scanAccountSubscriptionPlan(scan database.Scanner, includeCounts bool) (plan *types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
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
func (q *MariaDB) scanPlans(rows database.ResultIterator, includeCounts bool) (plans []*types.AccountSubscriptionPlan, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := q.scanAccountSubscriptionPlan(rows, includeCounts)
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
		q.logger.Error(closeErr, "closing database rows")
		return nil, 0, 0, closeErr
	}

	return plans, filteredCount, totalCount, nil
}

// buildGetPlanQuery constructs a SQL query for fetching an plan with a given ID belong to a user with a given ID.
func (q *MariaDB) buildGetPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.AccountSubscriptionPlansTableColumns...).
		From(queriers.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountSubscriptionPlansTableName, queriers.IDColumn):         planID,
			fmt.Sprintf("%s.%s", queriers.AccountSubscriptionPlansTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAccountSubscriptionPlan fetches an plan from the database.
func (q *MariaDB) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*types.AccountSubscriptionPlan, error) {
	query, args := q.buildGetPlanQuery(planID)
	row := q.db.QueryRowContext(ctx, query, args...)

	plan, _, _, err := q.scanAccountSubscriptionPlan(row, false)

	return plan, err
}

// buildGetAllPlansCountQuery returns a query that fetches the total number of plans in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *MariaDB) buildGetAllPlansCountQuery() string {
	allPlansCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AccountSubscriptionPlansTableName)).
		From(queriers.AccountSubscriptionPlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountSubscriptionPlansTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allPlansCountQuery
}

// GetAllAccountSubscriptionPlansCount will fetch the count of plans from the database.
func (q *MariaDB) GetAllAccountSubscriptionPlansCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllPlansCountQuery()).Scan(&count)
	return count, err
}

// buildGetPlansQuery builds a SQL query selecting plans that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *MariaDB) buildGetPlansQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		queriers.AccountSubscriptionPlansTableName,
		"",
		queriers.AccountSubscriptionPlansTableColumns,
		0,
		true,
		filter,
	)
}

// GetAccountSubscriptionPlans fetches a list of plans from the database that meet a particular filter.
func (q *MariaDB) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (*types.AccountSubscriptionPlanList, error) {
	query, args := q.buildGetPlansQuery(filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for plans: %w", err)
	}

	plans, filteredCount, totalCount, err := q.scanPlans(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AccountSubscriptionPlanList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: filteredCount,
			TotalCount:    totalCount,
		},
		AccountSubscriptionPlans: plans,
	}

	return list, nil
}

// buildCreatePlanQuery takes an plan and returns a creation query for that plan and the relevant arguments.
func (q *MariaDB) buildCreatePlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.AccountSubscriptionPlansTableName).
		Columns(
			queriers.AccountSubscriptionPlansTableNameColumn,
			queriers.AccountSubscriptionPlansTableDescriptionColumn,
			queriers.AccountSubscriptionPlansTablePriceColumn,
			queriers.AccountSubscriptionPlansTablePeriodColumn,
		).
		Values(
			input.Name,
			input.Description,
			input.Price,
			input.Period.String(),
		).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreateAccountSubscriptionPlan creates an plan in the database.
func (q *MariaDB) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	x := &types.AccountSubscriptionPlan{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Period:      input.Period,
	}

	query, args := q.buildCreatePlanQuery(x)

	// create the plan.
	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing plan creation query: %w", err)
	}

	x.CreatedOn = q.timeTeller.Now()
	x.ID = q.getIDFromResult(res)

	return x, nil
}

// buildUpdatePlanQuery takes an plan and returns an update SQL query, with the relevant query parameters.
func (q *MariaDB) buildUpdatePlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountSubscriptionPlansTableName).
		Set(queriers.AccountSubscriptionPlansTableNameColumn, input.Name).
		Set(queriers.AccountSubscriptionPlansTableDescriptionColumn, input.Description).
		Set(queriers.AccountSubscriptionPlansTablePriceColumn, input.Price).
		Set(queriers.AccountSubscriptionPlansTablePeriodColumn, input.Period.String()).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn: input.ID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateAccountSubscriptionPlan updates a particular plan. Note that UpdatePlan expects the provided input to have a valid ID.
func (q *MariaDB) UpdateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlan) error {
	query, args := q.buildUpdatePlanQuery(input)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchivePlanQuery returns a SQL query which marks a given plan belonging to a given user as archived.
func (q *MariaDB) buildArchivePlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountSubscriptionPlansTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:         planID,
			queriers.ArchivedOnColumn: nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveAccountSubscriptionPlan marks an plan as archived in the database.
func (q *MariaDB) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	query, args := q.buildArchivePlanQuery(planID)

	res, err := q.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogAccountSubscriptionPlanCreationEvent saves a AccountSubscriptionPlanCreationEvent in the audit log table.
func (q *MariaDB) LogAccountSubscriptionPlanCreationEvent(ctx context.Context, plan *types.AccountSubscriptionPlan) {
	q.createAuditLogEntry(ctx, audit.BuildAccountSubscriptionPlanCreationEventEntry(plan))
}

// AccountSubscriptionLogPlanUpdateEvent saves a AccountSubscriptionPlanUpdateEvent in the audit log table.
func (q *MariaDB) AccountSubscriptionLogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildAccountSubscriptionPlanUpdateEventEntry(userID, planID, changes))
}

// AccountSubscriptionLogPlanArchiveEvent saves a AccountSubscriptionPlanArchiveEvent in the audit log table.
func (q *MariaDB) AccountSubscriptionLogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildAccountSubscriptionPlanArchiveEventEntry(userID, planID))
}

// buildGetAuditLogEntriesForPlanQuery constructs a SQL query for fetching audit log entries
// associated with a given plan.
func (q *MariaDB) buildGetAuditLogEntriesForPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Expr(
			fmt.Sprintf(
				jsonPluckQuery,
				queriers.AuditLogEntriesTableName,
				queriers.AuditLogEntriesTableContextColumn,
				planID,
				audit.AccountSubscriptionPlanAssignmentKey,
			),
		)).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForAccountSubscriptionPlan fetches a audit log entries for a given plan from the database.
func (q *MariaDB) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]*types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForPlanQuery(planID)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}
