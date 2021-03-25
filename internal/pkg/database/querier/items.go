package querier

import (
	"context"
	"database/sql"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.ItemDataManager = (*Client)(nil)
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (c *Client) scanItem(ctx context.Context, scan database.Scanner, includeCounts bool) (x *types.Item, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	x = &types.Item{}

	targetVars := []interface{}{
		&x.ID,
		&x.ExternalID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToAccount,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "")
	}

	return x, filteredCount, totalCount, nil
}

// scanItems takes some database rows and turns them into a slice of items.
func (c *Client) scanItems(ctx context.Context, rows database.ResultIterator, includeCounts bool) (items []*types.Item, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		x, fc, tc, scanErr := c.scanItem(ctx, rows, includeCounts)
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

	if err = c.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return items, filteredCount, totalCount, nil
}

// ItemExists fetches whether or not an item exists from the database.
func (c *Client) ItemExists(ctx context.Context, itemID, accountID uint64) (exists bool, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	query, args := c.sqlQueryBuilder.BuildItemExistsQuery(itemID, accountID)

	return c.performBooleanQuery(ctx, c.db, query, args)
}

// GetItem fetches an item from the database.
func (c *Client) GetItem(ctx context.Context, itemID, accountID uint64) (*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.ItemIDKey: itemID,
		keys.UserIDKey: accountID,
	})

	query, args := c.sqlQueryBuilder.BuildGetItemQuery(itemID, accountID)
	row := c.db.QueryRowContext(ctx, query, args...)

	item, _, _, err := c.scanItem(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning item")
	}

	return item, nil
}

// GetAllItemsCount fetches the count of items from the database that meet a particular filter.
func (c *Client) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllItemsCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllItemsCountQuery())
}

// GetAllItems fetches a list of all items in the database.
func (c *Client) GetAllItems(ctx context.Context, results chan []*types.Item, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("batch_size", batchSize)

	count, err := c.GetAllItemsCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of items")
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfItemsQuery(begin, end)
			logger = logger.WithValues(map[string]interface{}{
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

			items, _, _, scanErr := c.scanItems(ctx, rows, false)
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
func (c *Client) GetItems(ctx context.Context, accountID uint64, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.ItemList{}
	logger := filter.AttachToLogger(c.logger).WithValue(keys.AccountIDKey, accountID)

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachQueryFilterToSpan(span, filter)

	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetItemsQuery(accountID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing items list retrieval query")
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = c.scanItems(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return x, nil
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (c *Client) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(c.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.ItemList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetItemsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing items list retrieval query for admin")
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = c.scanItems(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return x, nil
}

// GetItemsWithIDs fetches items from the database within a given set of IDs.
func (c *Client) GetItemsWithIDs(ctx context.Context, accountID uint64, limit uint8, ids []uint64) ([]*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)

	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	logger := c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: accountID,
		"limit":        limit,
		"id_count":     len(ids),
	})

	query, args := c.sqlQueryBuilder.BuildGetItemsWithIDsQuery(accountID, limit, ids, false)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching items from database")
	}

	items, _, _, err := c.scanItems(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return items, nil
}

// GetItemsWithIDsForAdmin fetches items from the database within a given set of IDs.
func (c *Client) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	logger := c.logger.WithValues(map[string]interface{}{
		"limit":    limit,
		"id_count": len(ids),
		"ids":      ids,
	})

	query, args := c.sqlQueryBuilder.BuildGetItemsWithIDsQuery(0, limit, ids, true)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching items from database")
	}

	items, _, _, err := c.scanItems(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return items, nil
}

// CreateItem creates an item in the database.
func (c *Client) CreateItem(ctx context.Context, input *types.ItemCreationInput, createdByUser uint64) (*types.Item, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.RequesterKey, createdByUser)

	logger.Debug("CreateItem called")

	query, args := c.sqlQueryBuilder.BuildCreateItemQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the item.
	id, err := c.performWriteQuery(ctx, tx, false, "item creation", query, args)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating item")
	}

	x := &types.Item{
		ID:               id,
		Name:             input.Name,
		Details:          input.Details,
		BelongsToAccount: input.BelongsToAccount,
		CreatedOn:        c.currentTime(),
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildItemCreationEventEntry(x, createdByUser)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing item creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	return x, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the
// provided input to have a valid ID.
func (c *Client) UpdateItem(ctx context.Context, updated *types.Item, changedByUser uint64, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.ItemIDKey, updated.ID)

	logger.Debug("UpdateItem called")
	tracing.AttachItemIDToSpan(span, updated.ID)
	tracing.AttachUserIDToSpan(span, changedByUser)

	query, args := c.sqlQueryBuilder.BuildUpdateItemQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "item update", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating item")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildItemUpdateEventEntry(changedByUser, updated.ID, updated.BelongsToAccount, changes)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing item update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// ArchiveItem archives an item from the database by its ID.
func (c *Client) ArchiveItem(ctx context.Context, itemID, belongsToAccount, archivedBy uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, belongsToAccount)
	tracing.AttachUserIDToSpan(span, archivedBy)
	tracing.AttachItemIDToSpan(span, itemID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.ItemIDKey:    itemID,
		keys.UserIDKey:    archivedBy,
		keys.AccountIDKey: belongsToAccount,
	})

	logger.Debug("ArchiveItem called")

	query, args := c.sqlQueryBuilder.BuildArchiveItemQuery(itemID, belongsToAccount)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "item archive", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating item")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildItemArchiveEventEntry(archivedBy, belongsToAccount, itemID)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing item archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForItem fetches a list of audit log entries from the database that relate to a given item.
func (c *Client) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.ItemIDKey, itemID)

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForItemQuery(itemID)

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
