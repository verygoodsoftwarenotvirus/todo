package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.ItemDataManager = (*Postgres)(nil)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (q *Postgres) scanItem(scan database.Scanner, includeCount bool) (*types.Item, uint64, error) {
	var (
		x     = &types.Item{}
		count uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	return x, count, nil
}

// scanItems takes some database rows and turns them into a slice of items.
func (q *Postgres) scanItems(rows database.ResultIterator, includeCount bool) ([]types.Item, uint64, error) {
	var (
		list  []types.Item
		count uint64
	)

	for rows.Next() {
		x, c, err := q.scanItem(rows, includeCount)
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

// buildItemExistsQuery constructs a SQL query for checking if an item with a given ID belong to a user with a given ID exists.
func (q *Postgres) buildItemExistsQuery(itemID, userID uint64) (string, []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn)).
		Prefix(queriers.ExistencePrefix).
		From(queriers.ItemsTableName).
		Suffix(queriers.ExistenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
		}),
	)
}

// ItemExists queries the database to see if a given item belonging to a given user exists.
func (q *Postgres) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	query, args := q.buildItemExistsQuery(itemID, userID)

	err = q.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

// buildGetItemQuery constructs a SQL query for fetching an item with a given ID belong to a user with a given ID.
func (q *Postgres) buildGetItemQuery(itemID, userID uint64) (string, []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn):                      itemID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn):              nil,
		}),
	)
}

// GetItem fetches an item from the database.
func (q *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*types.Item, error) {
	query, args := q.buildGetItemQuery(itemID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	item, _, err := q.scanItem(row, false)

	return item, err
}

// buildGetAllItemsCountQuery returns a query that fetches the total number of items in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Postgres) buildGetAllItemsCountQuery() string {
	query, _ := q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.ItemsTableName)).
		From(queriers.ItemsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
		}),
	)

	return query
}

// GetAllItemsCount will fetch the count of items from the database.
func (q *Postgres) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfItemsQuery returns a query that fetches every item in the database within a bucketed range.
func (q *Postgres) buildGetBatchOfItemsQuery(beginID, endID uint64) (string, []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.IDColumn): endID,
		}),
	)
}

// GetAllItems fetches every item from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (q *Postgres) GetAllItems(ctx context.Context, resultChannel chan []types.Item) error {
	count, err := q.GetAllItemsCount(ctx)
	if err != nil {
		return fmt.Errorf("error fetching count of items: %w", err)
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := q.buildGetBatchOfItemsQuery(begin, end)
			logger := q.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := q.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			items, _, err := q.scanItems(rows, false)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- items
		}(beginID, endID)
	}

	return nil
}

// buildGetItemsQuery builds a SQL query selecting items that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) buildGetItemsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (string, []interface{}) {
	where := squirrel.Eq{}
	if forAdmin {
		if filter != nil && filter.IncludeArchived {
			where[fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn)] = nil
		}
	} else {
		where[fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn)] = nil
		where[fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn)] = userID
	}

	countQueryBuilder := q.sqlBuilder.PlaceholderFormat(squirrel.Question).
		Select(allCountQuery).
		From(queriers.ItemsTableName).
		Where(where)

	if filter != nil {
		countQueryBuilder = queriers.ApplyFilterToSubCountQueryBuilder(filter, countQueryBuilder, queriers.ItemsTableName)
	}

	countQuery, countQueryArgs := q.buildQuery(countQueryBuilder)

	builder := q.sqlBuilder.
		Select(append(queriers.ItemsTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.ItemsTableName).
		Where(where).
		OrderBy(fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.CreatedOnColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.ItemsTableName)
	}

	query, selectArgs := q.buildQuery(builder)

	return query, append(countQueryArgs, selectArgs...)
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (q *Postgres) GetItems(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := q.buildGetItemsQuery(userID, false, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for items: %w", err)
	}

	items, count, err := q.scanItems(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Items: items,
	}

	return list, nil
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (q *Postgres) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.ItemList, error) {
	query, args := q.buildGetItemsQuery(0, true, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for items: %w", err)
	}

	items, count, err := q.scanItems(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.ItemList{
		Pagination: types.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Items: items,
	}

	return list, nil
}

// buildGetItemsWithIDsQuery builds a SQL query selecting items that belong to a given user,
// and have IDs that exist within a given set of IDs. Returns both the query and the relevant
// args to pass to the query executor. This function is primarily intended for use with a search
// index, which would provide a slice of string IDs to query against. This function accepts a
// slice of uint64s instead of a slice of strings in order to ensure all the provided strings
// are valid database IDs, because there's no way in squirrel to escape them in the unnest join,
// and if we accept strings we could leave ourselves vulnerable to SQL injection attacks.
func (q *Postgres) buildGetItemsWithIDsQuery(userID uint64, limit uint8, ids []uint64, forAdmin bool) (string, []interface{}) {
	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ArchivedOnColumn): nil,
	}
	if !forAdmin {
		where[fmt.Sprintf("%s.%s", queriers.ItemsTableName, queriers.ItemsTableUserOwnershipColumn)] = userID
	}

	subqueryBuilder := q.sqlBuilder.Select(queriers.ItemsTableColumns...).
		From(queriers.ItemsTableName).
		Join(fmt.Sprintf("unnest('{%s}'::int[])", joinUint64s(ids))).
		Suffix(fmt.Sprintf("WITH ORDINALITY t(id, ord) USING (id) ORDER BY t.ord LIMIT %d", limit))
	builder := q.sqlBuilder.
		Select(queriers.ItemsTableColumns...).
		FromSelect(subqueryBuilder, queriers.ItemsTableName).
		Where(where)

	return q.buildQuery(builder)
}

// GetItemsWithIDs fetches a list of items from the database that exist within a given set of IDs.
func (q *Postgres) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := q.buildGetItemsWithIDsQuery(userID, limit, ids, false)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for items: %w", err)
	}

	items, _, err := q.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// GetItemsWithIDsForAdmin fetches a list of items from the database that exist within a given set of IDs.
func (q *Postgres) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]types.Item, error) {
	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := q.buildGetItemsWithIDsQuery(0, limit, ids, true)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for items: %w", err)
	}

	items, _, err := q.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return items, nil
}

// buildCreateItemQuery takes an item and returns a creation query for that item and the relevant arguments.
func (q *Postgres) buildCreateItemQuery(input *types.Item) (string, []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(queriers.ItemsTableName).
		Columns(
			queriers.ItemsTableNameColumn,
			queriers.ItemsTableDetailsColumn,
			queriers.ItemsTableUserOwnershipColumn,
		).
		Values(
			input.Name,
			input.Details,
			input.BelongsToUser,
		).
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)),
	)
}

// CreateItem creates an item in the database.
func (q *Postgres) CreateItem(ctx context.Context, input *types.ItemCreationInput) (*types.Item, error) {
	x := &types.Item{
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := q.buildCreateItemQuery(x)

	// create the item.
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, fmt.Errorf("error executing item creation query: %w", err)
	}

	return x, nil
}

// buildUpdateItemQuery takes an item and returns an update SQL query, with the relevant query parameters.
func (q *Postgres) buildUpdateItemQuery(input *types.Item) (string, []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.ItemsTableNameColumn, input.Name).
		Set(queriers.ItemsTableDetailsColumn, input.Details).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      input.ID,
			queriers.ItemsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)),
	)
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (q *Postgres) UpdateItem(ctx context.Context, input *types.Item) error {
	query, args := q.buildUpdateItemQuery(input)
	return q.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// buildArchiveItemQuery returns a SQL query which marks a given item belonging to a given user as archived.
func (q *Postgres) buildArchiveItemQuery(itemID, userID uint64) (string, []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(queriers.ItemsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                      itemID,
			queriers.ArchivedOnColumn:              nil,
			queriers.ItemsTableUserOwnershipColumn: userID,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.ArchivedOnColumn)),
	)
}

// ArchiveItem marks an item as archived in the database.
func (q *Postgres) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	query, args := q.buildArchiveItemQuery(itemID, userID)

	res, err := q.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogItemCreationEvent saves a ItemCreationEvent in the audit log table.
func (q *Postgres) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	q.createAuditLogEntry(ctx, audit.BuildItemCreationEventEntry(item))
}

// LogItemUpdateEvent saves a ItemUpdateEvent in the audit log table.
func (q *Postgres) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildItemUpdateEventEntry(userID, itemID, changes))
}

// LogItemArchiveEvent saves a ItemArchiveEvent in the audit log table.
func (q *Postgres) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildItemArchiveEventEntry(userID, itemID))
}

// buildGetAuditLogEntriesForItemQuery constructs a SQL query for fetching audit log entries
// associated with a given item.
func (q *Postgres) buildGetAuditLogEntriesForItemQuery(itemID uint64) (query string, args []interface{}) {
	itemIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.ItemAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{itemIDKey: itemID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	return q.buildQuery(builder)
}

// GetAuditLogEntriesForItem fetches a audit log entries for a given item from the database.
func (q *Postgres) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForItemQuery(itemID)

	q.logger.WithValues(map[string]interface{}{
		"item_id": itemID,
		"query":   query,
	}).Debug("GetAuditLogEntriesForItem called")

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
