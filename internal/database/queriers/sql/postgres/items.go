package postgres

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
)

var (
	_ types.ItemDataManager = (*SQLQuerier)(nil)

	// itemsTableColumns are the columns for the items table.
	itemsTableColumns = []string{
		"items.id",
		"items.name",
		"items.details",
		"items.created_on",
		"items.last_updated_on",
		"items.archived_on",
		"items.belongs_to_account",
	}
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

const itemExistenceQuery = "SELECT EXISTS ( SELECT items.id FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1 AND items.id = $2 )"

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

	args := []interface{}{
		accountID,
		itemID,
	}

	result, err := q.performBooleanQuery(ctx, q.db, itemExistenceQuery, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing item existence check")
	}

	return result, nil
}

const getItemQuery = `
	SELECT items.id, 
items.name, 
items.details, 
items.created_on, 
items.last_updated_on, 
items.archived_on, 
items.belongs_to_account FROM items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1 AND items.id = $2
`

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

	args := []interface{}{
		accountID,
		itemID,
	}

	row := q.getOneRow(ctx, q.db, "item", getItemQuery, args)

	item, _, _, err := q.scanItem(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning item")
	}

	return item, nil
}

const getAllItemsCountQuery = "SELECT COUNT(items.id) FROM items WHERE items.archived_on IS NULL"

// GetTotalItemCount fetches the count of items from the database that meet a particular filter.
func (q *SQLQuerier) GetTotalItemCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, getAllItemsCountQuery, "fetching count of items")
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

	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", querybuilding.ItemsTableName, querybuilding.ArchivedOnColumn): nil,
	}

	query, args := q.buildListQuery(
		ctx,
		"items",
		nil,
		where,
		"belongs_to_account",
		itemsTableColumns,
		accountID,
		false,
		filter,
	)

	rows, err := q.performReadQuery(ctx, q.db, "items", query, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing items list retrieval query")
	}

	if x.Items, x.FilteredCount, x.TotalCount, err = q.scanItems(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning items")
	}

	return x, nil
}

const getItemsWithIDsQuery = `
	SELECT items.id, 
items.name, 
items.details, 
items.created_on, 
items.last_updated_on, 
items.archived_on, 
items.belongs_to_account FROM (SELECT items.id, 
items.name, 
items.details, 
items.created_on, 
items.last_updated_on, 
items.archived_on, 
items.belongs_to_account FROM items JOIN unnest('{%s}'::text[]) WITH ORDINALITY t(id, ord) USING (id) ORDER BY t.ord LIMIT 20) AS items WHERE items.archived_on IS NULL AND items.belongs_to_account = $1 AND items.id IN ($2,$3,$4)
`

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

	query := fmt.Sprintf(getItemsWithIDsQuery, joinIDs(ids))

	args := []interface{}{accountID}
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := q.performReadQuery(ctx, q.db, "items with IDs", query, args)
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
func (q *SQLQuerier) CreateItem(ctx context.Context, input *types.ItemDatabaseCreationInput) (*types.Item, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.ItemIDKey, input.ID)

	query := `
		INSERT INTO items (id,name,details,belongs_to_account) VALUES ($1,$2,$3,$4)
	`
	args := []interface{}{
		input.ID,
		input.Name,
		input.Details,
		input.BelongsToAccount,
	}

	// create the item.
	if err := q.performWriteQuery(ctx, q.db, "item creation", query, args); err != nil {
		return nil, observability.PrepareError(err, logger, span, "creating item")
	}

	x := &types.Item{
		ID:               input.ID,
		Name:             input.Name,
		Details:          input.Details,
		BelongsToAccount: input.BelongsToAccount,
		CreatedOn:        q.currentTime(),
	}

	tracing.AttachItemIDToSpan(span, x.ID)
	logger.Info("item created")

	return x, nil
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (q *SQLQuerier) UpdateItem(ctx context.Context, updated *types.Item) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.ItemIDKey, updated.ID)
	tracing.AttachItemIDToSpan(span, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.BelongsToAccount)

	updateItemQuery := `
		UPDATE items SET name = $1, details = $2, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $3 AND id = $4
	`
	args := []interface{}{
		updated.Name,
		updated.Details,
		updated.BelongsToAccount,
		updated.ID,
	}

	if err := q.performWriteQuery(ctx, q.db, "item update", updateItemQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "updating item")
	}

	logger.Info("item updated")

	return nil
}

const archiveItemQuery = `
	UPDATE items SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_account = $1 AND id = $2
`

// ArchiveItem archives an item from the database by its ID.
func (q *SQLQuerier) ArchiveItem(ctx context.Context, itemID, accountID string) error {
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

	args := []interface{}{
		accountID,
		itemID,
	}

	if err := q.performWriteQuery(ctx, q.db, "item archive", archiveItemQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "updating item")
	}

	logger.Info("item archived")

	return nil
}
