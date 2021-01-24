package superclient

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.ItemDataManager  = (*Client)(nil)
	_ types.ItemAuditManager = (*Client)(nil)
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (c *Client) scanItem(scan database.Scanner, includeCounts bool) (x *types.Item, filteredCount, totalCount uint64, err error) {
	x = &types.Item{}

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	return x, filteredCount, totalCount, nil
}

// scanItems takes some database rows and turns them into a slice of items.
func (c *Client) scanItems(rows database.ResultIterator, includeCounts bool) (items []*types.Item, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := c.scanItem(rows, includeCounts)
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

		items = append(items, x)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing database rows")
		return nil, 0, 0, closeErr
	}

	return items, filteredCount, totalCount, nil
}

// ItemExists fetches whether or not an item exists from the database.
func (c *Client) ItemExists(ctx context.Context, itemID, userID uint64) (exists bool, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.ItemIDKey: itemID,
		keys.UserIDKey: userID,
	}).Debug("ItemExists called")

	query, args := c.sqlQueryBuilder.BuildItemExistsQuery(itemID, userID)

	if err = c.db.QueryRowContext(ctx, query, args...).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("executing item existence query: %w", err)
	}

	return exists, nil
}

// GetItem fetches an item from the database.
func (c *Client) GetItem(ctx context.Context, itemID, userID uint64) (*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.ItemIDKey: itemID,
		keys.UserIDKey: userID,
	}).Debug("GetItem called")

	query, args := c.sqlQueryBuilder.BuildGetItemQuery(itemID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	item, _, _, err := c.scanItem(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning item: %w", err)
	}

	return item, nil
}

// GetAllItemsCount fetches the count of items from the database that meet a particular filter.
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	if err = c.db.QueryRowContext(ctx, c.sqlQueryBuilder.BuildGetAllItemsCountQuery()).Scan(&count); err != nil {
		return 0, fmt.Errorf("executing items count query: %w", err)
	}

	return count, nil
}

// GetAllItems fetches a list of all items in the database.
func (c *Client) GetAllItems(ctx context.Context, results chan []*types.Item, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllItems called")

	count, countErr := c.GetAllItemsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of items: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfItemsQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			items, _, _, scanErr := c.scanItems(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			results <- items
		}(beginID, endID)
	}

	return nil
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (c *Client) GetItems(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.ItemList{}

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("GetItems called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetItemsQuery(userID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing items list retrieval query: %w", err)
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = c.scanItems(rows, true); err != nil {
		return nil, fmt.Errorf("scanning items: %w", err)
	}

	return x, nil
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (c *Client) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.ItemList{}

	c.logger.Debug("GetItemsForAdmin called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetItemsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing items list retrieval query for admin: %w", err)
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = c.scanItems(rows, true); err != nil {
		return nil, fmt.Errorf("scanning items: %w", err)
	}

	return x, nil
}

// GetItemsWithIDs fetches items from the database within a given set of IDs.
func (c *Client) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
		"limit":        limit,
		"id_count":     len(ids),
	}).Debug("GetItemsWithIDs called")

	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	query, args := c.sqlQueryBuilder.BuildGetItemsWithIDsQuery(userID, limit, ids, false)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, _, _, err := c.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning items: %w", err)
	}

	return items, nil
}

// GetItemsWithIDsForAdmin fetches items from the database within a given set of IDs.
func (c *Client) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValues(map[string]interface{}{
		"limit":    limit,
		"id_count": len(ids),
		"ids":      ids,
	}).Debug(fmt.Sprintf("GetItemsWithIDsForAdmin called: %v", ids))

	query, args := c.sqlQueryBuilder.BuildGetItemsWithIDsQuery(0, limit, ids, true)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	items, _, _, err := c.scanItems(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning items: %w", err)
	}

	return items, nil
}

// CreateItem creates an item in the database.
func (c *Client) CreateItem(ctx context.Context, input *types.ItemCreationInput) (*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateItem called")

	query, args := c.sqlQueryBuilder.BuildCreateItemQuery(input)

	// create the item.
	res, err := c.execContextAndReturnResult(ctx, "item creation", query, args...)
	if err != nil {
		return nil, err
	}

	x := &types.Item{
		ID:            c.getIDFromResult(res),
		Name:          input.Name,
		Details:       input.Details,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	return x, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the
// provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, updated *types.Item) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.ItemIDKey, updated.ID).Debug("UpdateItem called")

	query, args := c.sqlQueryBuilder.BuildUpdateItemQuery(updated)

	return c.execContext(ctx, "item update", query, args...)
}

// ArchiveItem archives an item from the database by its ID.
func (c *Client) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachItemIDToSpan(span, itemID)

	c.logger.WithValues(map[string]interface{}{
		"item_id":      itemID,
		keys.UserIDKey: userID,
	}).Debug("ArchiveItem called")

	query, args := c.sqlQueryBuilder.BuildArchiveItemQuery(itemID, userID)

	return c.execContext(ctx, "item archive", query, args...)
}

// LogItemCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, item.BelongsToUser).Debug("LogItemCreationEvent called")

	c.createAuditLogEntry(ctx, audit.BuildItemCreationEventEntry(item))
}

// LogItemUpdateEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogItemUpdateEvent called")

	c.createAuditLogEntry(ctx, audit.BuildItemUpdateEventEntry(userID, itemID, changes))
}

// LogItemArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogItemArchiveEvent called")

	c.createAuditLogEntry(ctx, audit.BuildItemArchiveEventEntry(userID, itemID))
}

// GetAuditLogEntriesForItem fetches a list of audit log entries from the database that relate to a given item.
func (c *Client) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForItem called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForItemQuery(itemID)

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
