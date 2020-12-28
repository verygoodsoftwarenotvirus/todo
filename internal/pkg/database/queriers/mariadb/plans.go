package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.PlanDataManager = (*MariaDB)(nil)

// scanPlan takes a database Scanner (i.e. *sql.Row) and scans the result into an Plan struct.
func (q *MariaDB) scanPlan(scan database.Scanner, includeCount bool) (*types.Plan, uint64, error) {
	var (
		x     = &types.Plan{}
		count uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Description,
		&x.Price,
		&x.Period,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	return x, count, nil
}

// scanPlans takes a logger and some database rows and turns them into a slice of plans.
func (q *MariaDB) scanPlans(rows database.ResultIterator, includeCount bool) ([]types.Plan, uint64, error) {
	var (
		list  []types.Plan
		count uint64
	)

	for rows.Next() {
		x, c, err := q.scanPlan(rows, includeCount)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 && includeCount {
			count = c
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		q.logger.Error(closeErr, "closing database rows")
	}

	return list, count, nil
}

// buildGetPlanQuery constructs a SQL query for fetching an plan with a given ID belong to a user with a given ID.
func (q *MariaDB) buildGetPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.PlansTableColumns...).
		From(queriers.PlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.PlansTableName, queriers.IDColumn): planID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetPlan fetches an plan from the database.
func (q *MariaDB) GetPlan(ctx context.Context, planID uint64) (*types.Plan, error) {
	query, args := q.buildGetPlanQuery(planID)
	row := q.db.QueryRowContext(ctx, query, args...)

	plan, _, err := q.scanPlan(row, false)

	return plan, err
}

// buildGetAllPlansCountQuery returns a query that fetches the total number of plans in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *MariaDB) buildGetAllPlansCountQuery() string {
	allPlansCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.PlansTableName)).
		From(queriers.PlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.PlansTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allPlansCountQuery
}

// GetAllPlansCount will fetch the count of plans from the database.
func (q *MariaDB) GetAllPlansCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllPlansCountQuery()).Scan(&count)
	return count, err
}

// buildGetPlansQuery builds a SQL query selecting plans that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *MariaDB) buildGetPlansQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	countQueryBuilder := q.sqlBuilder.PlaceholderFormat(squirrel.Question).
		Select(allCountQuery).
		From(queriers.PlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.PlansTableName, queriers.ArchivedOnColumn): nil,
		})

	if filter != nil {
		countQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, countQueryBuilder, queriers.PlansTableName)
	}

	countQuery, countQueryArgs, err := countQueryBuilder.ToSql()
	q.logQueryBuildingError(err)

	builder := q.sqlBuilder.
		Select(append(queriers.PlansTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.PlansTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.PlansTableName, queriers.ArchivedOnColumn): nil,
		}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.PlansTableName, queriers.IDColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.PlansTableName)
	}

	query, selectArgs, err := builder.ToSql()
	q.logQueryBuildingError(err)

	return query, append(countQueryArgs, selectArgs...)
}

// GetPlans fetches a list of plans from the database that meet a particular filter.
func (q *MariaDB) GetPlans(ctx context.Context, filter *types.QueryFilter) (*types.PlanList, error) {
	query, args := q.buildGetPlansQuery(filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for plans: %w", err)
	}

	plans, count, err := q.scanPlans(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.PlanList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Plans: plans,
	}

	return list, nil
}

// buildCreatePlanQuery takes an plan and returns a creation query for that plan and the relevant arguments.
func (q *MariaDB) buildCreatePlanQuery(input *types.Plan) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.PlansTableName).
		Columns(
			queriers.PlansTableNameColumn,
		).
		Values(
			input.Name,
		).
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreatePlan creates an plan in the database.
func (q *MariaDB) CreatePlan(ctx context.Context, input *types.PlanCreationInput) (*types.Plan, error) {
	x := &types.Plan{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Period:      input.Period,
	}

	query, args := q.buildCreatePlanQuery(x)

	// create the plan.
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, fmt.Errorf("error executing plan creation query: %w", err)
	}

	return x, nil
}

// buildUpdatePlanQuery takes an plan and returns an update SQL query, with the relevant query parameters.
func (q *MariaDB) buildUpdatePlanQuery(input *types.Plan) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.PlansTableName).
		Set(queriers.PlansTableNameColumn, input.Name).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn: input.ID,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdatePlan updates a particular plan. Note that UpdatePlan expects the provided input to have a valid ID.
func (q *MariaDB) UpdatePlan(ctx context.Context, input *types.Plan) error {
	query, args := q.buildUpdatePlanQuery(input)
	return q.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// buildArchivePlanQuery returns a SQL query which marks a given plan belonging to a given user as archived.
func (q *MariaDB) buildArchivePlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.PlansTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:         planID,
			queriers.ArchivedOnColumn: nil,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.ArchivedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchivePlan marks an plan as archived in the database.
func (q *MariaDB) ArchivePlan(ctx context.Context, planID uint64) error {
	query, args := q.buildArchivePlanQuery(planID)

	res, err := q.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogPlanCreationEvent saves a PlanCreationEvent in the audit log table.
func (q *MariaDB) LogPlanCreationEvent(ctx context.Context, plan *types.Plan) {
	q.createAuditLogEntry(ctx, audit.BuildPlanCreationEventEntry(plan))
}

// LogPlanUpdateEvent saves a PlanUpdateEvent in the audit log table.
func (q *MariaDB) LogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildPlanUpdateEventEntry(userID, planID, changes))
}

// LogPlanArchiveEvent saves a PlanArchiveEvent in the audit log table.
func (q *MariaDB) LogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildPlanArchiveEventEntry(userID, planID))
}

// buildGetAuditLogEntriesForPlanQuery constructs a SQL query for fetching audit log entries
// associated with a given plan.
func (q *MariaDB) buildGetAuditLogEntriesForPlanQuery(planID uint64) (query string, args []interface{}) {
	var err error

	planIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, planID, audit.PlanAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{planIDKey: planID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.IDColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForPlan fetches a audit log entries for a given plan from the database.
func (q *MariaDB) GetAuditLogEntriesForPlan(ctx context.Context, planID uint64) ([]types.AuditLogEntry, error) {
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
