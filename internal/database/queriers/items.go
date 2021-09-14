package queriers

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	_ types.ItemDataManager = (*SQLQuerier)(nil)
)

// scanItem takes a database Scanner (i.e. *sql.Row) and scans the result into an item struct.
func (q *SQLQuerier) scanItem(ctx context.Context, scan database.Scanner, includeCounts bool) (x *types.Item, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	x = &types.Item{}

	targetVars := []interface{}{
		&x.ID,
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

// ItemExists fetches whether an item exists from the database.
func (q *SQLQuerier) ItemExists(ctx context.Context, itemID, accountID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if itemID == "" {
		return false, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	if accountID == "" {
		return false, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	query, args := q.sqlQueryBuilder.BuildItemExistsQuery(ctx, itemID, accountID)

	result, err := q.performBooleanQuery(ctx, q.db, query, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing item existence check")
	}

	return result, nil
}

// GetItem fetches an item from the database.
func (q *SQLQuerier) GetItem(ctx context.Context, itemID, accountID string) (*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if itemID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	query, args := q.sqlQueryBuilder.BuildGetItemQuery(ctx, itemID, accountID)
	row := q.getOneRow(ctx, q.db, "item", query, args...)

	item, _, _, err := q.scanItem(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning item")
	}

	return item, nil
}

// GetTotalItemCount fetches the count of items from the database that meet a particular filter.
func (q *SQLQuerier) GetTotalItemCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetTotalItemCountQuery(ctx), "fetching count of items")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of items")
	}

	return count, nil
}

// GetItems fetches a list of items from the database that meet a particular filter.
func (q *SQLQuerier) GetItems(ctx context.Context, accountID string, filter *types.QueryFilter) (x *types.ItemList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	x = &types.ItemList{}
	logger = filter.AttachToLogger(logger)
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

// GetItemsWithIDs fetches items from the database within a given set of IDs.
func (q *SQLQuerier) GetItemsWithIDs(ctx context.Context, accountID string, limit uint8, ids []string) ([]*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if accountID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	if limit == 0 {
		limit = uint8(types.DefaultLimit)
	}

	logger = logger.WithValues(map[string]interface{}{
		"limit":    limit,
		"id_count": len(ids),
	})

	query, args := q.sqlQueryBuilder.BuildGetItemsWithIDsQuery(ctx, accountID, limit, ids, true)

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

// CreateItem creates an item in the database.
func (q *SQLQuerier) CreateItem(ctx context.Context, input *types.ItemDatabaseCreationInput, createdByUser string) (*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if createdByUser == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.RequesterIDKey, createdByUser)
	tracing.AttachRequestingUserIDToSpan(span, createdByUser)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildCreateItemQuery(ctx, input)

	// create the item.
	err = q.performWriteQueryIgnoringReturn(ctx, tx, "item creation", query, args)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating item")
	}

	x := &types.Item{
		ID:               input.ID,
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

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (q *SQLQuerier) UpdateItem(ctx context.Context, updated *types.Item, changedByUser string, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	if changedByUser == "" {
		return ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.ItemIDKey, updated.ID)
	tracing.AttachItemIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)
	tracing.AttachRequestingUserIDToSpan(span, changedByUser)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildUpdateItemQuery(ctx, updated)
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
func (q *SQLQuerier) ArchiveItem(ctx context.Context, itemID, accountID, archivedBy string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if itemID == "" {
		return ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
	tracing.AttachItemIDToSpan(span, itemID)

	if accountID == "" {
		return ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	if archivedBy == "" {
		return ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.RequesterIDKey, archivedBy)
	tracing.AttachUserIDToSpan(span, archivedBy)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildArchiveItemQuery(ctx, itemID, accountID)

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
func (q *SQLQuerier) GetAuditLogEntriesForItem(ctx context.Context, itemID string) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if itemID == "" {
		return nil, ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ItemIDKey, itemID)
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
