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
	_ types.ItemDataManager = (*SQLQuerier)(nil)
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an Item struct.
func (q *SQLQuerier) scanItem(ctx context.Context, scan database.Scanner, includeCounts bool) (x *types.Item, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

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
func (q *SQLQuerier) scanItems(ctx context.Context, rows database.ResultIterator, includeCounts bool) (items []*types.Item, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		x, fc, tc, scanErr := q.scanItem(ctx, rows, includeCounts)
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

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return items, filteredCount, totalCount, nil
}

// ItemExists fetches whether or not an item exists from the database.
func (q *SQLQuerier) ItemExists(ctx context.Context, itemID, accountID uint64) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return false, ErrInvalidIDProvided
	}

	if accountID == 0 {
		return false, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.ItemIDKey, itemID).WithValue(keys.AccountIDKey, accountID)
	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	query, args := q.sqlQueryBuilder.BuildItemExistsQuery(ctx, itemID, accountID)

	result, err := q.performBooleanQuery(ctx, q.db, query, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing item existence check")
	}

	return result, nil
}

// GetItem fetches an item from the database.
func (q *SQLQuerier) GetItem(ctx context.Context, itemID, accountID uint64) (*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachItemIDToSpan(span, itemID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.ItemIDKey: itemID,
		keys.UserIDKey: accountID,
	})

	query, args := q.sqlQueryBuilder.BuildGetItemQuery(ctx, itemID, accountID)
	row := q.getOneRow(ctx, q.db, "item", query, args...)

	item, _, _, err := q.scanItem(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning item")
	}

	return item, nil
}

// GetAllItemsCount fetches the count of items from the database that meet a particular filter.
func (q *SQLQuerier) GetAllItemsCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllItemsCountQuery(ctx), "fetching count of items")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of items")
	}

	return count, nil
}

// GetAllItems fetches a list of all items in the database.
func (q *SQLQuerier) GetAllItems(ctx context.Context, results chan []*types.Item, batchSize uint16) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if results == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue("batch_size", batchSize)

	count, err := q.GetAllItemsCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of items")
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := q.sqlQueryBuilder.BuildGetBatchOfItemsQuery(ctx, begin, end)
			logger = logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := q.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			items, _, _, scanErr := q.scanItems(ctx, rows, false)
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
func (q *SQLQuerier) GetItems(ctx context.Context, accountID uint64, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	x = &types.ItemList{}
	logger := filter.AttachToLogger(q.logger).WithValue(keys.AccountIDKey, accountID)

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachQueryFilterToSpan(span, filter)

	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetItemsQuery(ctx, accountID, false, filter)

	rows, err := q.performReadQuery(ctx, q.db, "items", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing items list retrieval query")
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = q.scanItems(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return x, nil
}

// GetItemsForAdmin fetches a list of items from the database that meet a particular filter for all users.
func (q *SQLQuerier) GetItemsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(q.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.ItemList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetItemsQuery(ctx, 0, true, filter)

	rows, err := q.performReadQuery(ctx, q.db, "items for admin", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing items list retrieval query for admin")
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = q.scanItems(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return x, nil
}

// GetItemsWithIDs fetches items from the database within a given set of IDs.
func (q *SQLQuerier) GetItemsWithIDs(ctx context.Context, accountID uint64, limit uint8, ids []uint64) ([]*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)

	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: accountID,
		"limit":        limit,
		"id_count":     len(ids),
	})

	query, args := q.sqlQueryBuilder.BuildGetItemsWithIDsQuery(ctx, accountID, limit, ids, false)

	rows, err := q.performReadQuery(ctx, q.db, "items with IDs", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching items from database")
	}

	items, _, _, err := q.scanItems(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return items, nil
}

// GetItemsWithIDsForAdmin fetches items from the database within a given set of IDs.
func (q *SQLQuerier) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	if len(ids) == 0 {
		return []*types.Item{}, nil
	}

	logger := q.logger.WithValues(map[string]interface{}{
		"limit":    limit,
		"id_count": len(ids),
		"ids":      ids,
	})

	query, args := q.sqlQueryBuilder.BuildGetItemsWithIDsQuery(ctx, 0, limit, ids, true)

	rows, err := q.performReadQuery(ctx, q.db, "items with IDs for admin", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching items from database")
	}

	items, _, _, err := q.scanItems(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return items, nil
}

// CreateItem creates an item in the database.
func (q *SQLQuerier) CreateItem(ctx context.Context, input *types.ItemCreationInput, createdByUser uint64) (*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.RequesterIDKey, createdByUser)
	tracing.AttachRequestingUserIDToSpan(span, createdByUser)

	query, args := q.sqlQueryBuilder.BuildCreateItemQuery(ctx, input)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the item.
	id, err := q.performWriteQuery(ctx, tx, false, "item creation", query, args)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating item")
	}

	x := &types.Item{
		ID:               id,
		Name:             input.Name,
		Details:          input.Details,
		BelongsToAccount: input.BelongsToAccount,
		CreatedOn:        q.currentTime(),
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildItemCreationEventEntry(x, createdByUser)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing item creation audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachItemIDToSpan(span, x.ID)
	logger.Info("item created")

	return x, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the
// provided input to have a valid ID.
func (q *SQLQuerier) UpdateItem(ctx context.Context, updated *types.Item, changedByUser uint64, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.ItemIDKey, updated.ID)
	tracing.AttachItemIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)
	tracing.AttachRequestingUserIDToSpan(span, changedByUser)

	query, args := q.sqlQueryBuilder.BuildUpdateItemQuery(ctx, updated)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "item update", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating item")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildItemUpdateEventEntry(changedByUser, updated.ID, updated.BelongsToAccount, changes)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing item update audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("item updated")

	return nil
}

// ArchiveItem archives an item from the database by its ID.
func (q *SQLQuerier) ArchiveItem(ctx context.Context, itemID, accountID, archivedBy uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return ErrInvalidIDProvided
	}

	if accountID == 0 {
		return ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, archivedBy)
	tracing.AttachItemIDToSpan(span, itemID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.ItemIDKey:    itemID,
		keys.UserIDKey:    archivedBy,
		keys.AccountIDKey: accountID,
	})

	query, args := q.sqlQueryBuilder.BuildArchiveItemQuery(ctx, itemID, accountID)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "item archive", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating item")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildItemArchiveEventEntry(archivedBy, accountID, itemID)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing item archive audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("item archived")

	return nil
}

// GetAuditLogEntriesForItem fetches a list of audit log entries from the database that relate to a given item.
func (q *SQLQuerier) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if itemID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntriesForItemQuery(ctx, itemID)

	rows, err := q.performReadQuery(ctx, q.db, "audit log entries for item", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entries")
	}

	return auditLogEntries, nil
}
