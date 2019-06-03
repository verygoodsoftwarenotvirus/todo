package postgres

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/Masterminds/squirrel"
)

const (
	itemsTableName = "items"
)

var (
	itemsTableColumns = []string{
		"id",
		"name",
		"details",
		"created_on",
		"updated_on",
		"archived_on",
		"belongs_to",
	}
)

func scanItem(scan database.Scanner) (*models.Item, error) {
	x := &models.Item{}

	if err := scan.Scan(
		&x.ID,
		&x.Name,
		&x.Details,
		&x.CreatedOn,
		&x.UpdatedOn,
		&x.ArchivedOn,
		&x.BelongsTo,
	); err != nil {
		return nil, err
	}

	return x, nil
}

func scanItems(logger logging.Logger, rows *sql.Rows) ([]models.Item, error) {
	var list []models.Item

	cols, err := rows.Columns()
	logger.
		WithError(err).
		WithValue("columns", strings.Join(cols, "|")).
		Debug("scanItems Called")

	for rows.Next() {
		x, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	logger.WithValue("row_count", len(list)).Debug("returning fine from scanItems")

	// RENAMEME this call proves this function's name is bad
	logQueryBuildingError(logger, rows.Close())

	return list, nil
}

func (p *Postgres) buildGetItemQuery(itemID, userID uint64) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq{
			"id":         itemID,
			"belongs_to": userID,
		}).ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetItem fetches an item from the postgres db
func (p *Postgres) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	query, args := p.buildGetItemQuery(itemID, userID)
	row := p.db.QueryRowContext(ctx, query, args...)
	return scanItem(row)
}

func (p *Postgres) buildGetItemCountQuery(filter *models.QueryFilter, userID uint64) (string, []interface{}) {
	builder := p.sqlBuilder.
		Select("COUNT(*)").
		From(itemsTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":  userID,
			"archived_on": nil,
		}))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err := builder.ToSql()
	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetItemCount will fetch the count of items from the postgres db that meet a particular filter and belong to a particular user.
func (p *Postgres) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (count uint64, err error) {
	query, args := p.buildGetItemCountQuery(filter, userID)
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (p *Postgres) buildGetAllItemsCountQuery() string {
	query, _, err := p.sqlBuilder.Select("COUNT(*)").
		From(itemsTableName).
		Where(squirrel.Eq{"archived_on": nil}).
		ToSql()
	logQueryBuildingError(p.logger, err)
	return query
}

// GetAllItemsCount will fetch the count of items from the postgres db that meet a particular filter
func (p *Postgres) GetAllItemsCount(ctx context.Context) (count uint64, err error) {
	err = p.db.QueryRowContext(ctx, p.buildGetAllItemsCountQuery()).Scan(&count)
	return count, err
}

func (p *Postgres) buildGetItemsQuery(filter *models.QueryFilter, userID uint64) (string, []interface{}) {
	builder := p.sqlBuilder.
		Select(itemsTableColumns...).
		From(itemsTableName).
		Where(squirrel.Eq(map[string]interface{}{
			"belongs_to":  userID,
			"archived_on": nil,
		}))

	if filter != nil {
		builder = filter.ApplyToQueryBuilder(builder)
	}

	query, args, err := builder.ToSql()
	logQueryBuildingError(p.logger, err)

	return query, args
}

// GetItems fetches a list of items from the postgres db that meet a particular filter
func (p *Postgres) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	query, args := p.buildGetItemsQuery(filter, userID)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	list, err := scanItems(p.logger, rows)
	if err != nil {
		return nil, err
	}

	count, err := p.GetItemCount(ctx, filter, userID)
	if err != nil {
		return nil, err
	}

	x := &models.ItemList{
		Pagination: models.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: count,
		},
		Items: list,
	}

	return x, nil
}

func (p *Postgres) buildCreateItemQuery(input *models.Item) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Insert(itemsTableName).
		Columns(
			"name",
			"details",
			"belongs_to",
		).
		Values(
			input.Name,
			input.Details,
			input.BelongsTo,
		).
		Suffix("RETURNING id, created_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// CreateItem creates an item in a postgres db
func (p *Postgres) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	i := &models.Item{
		Name:      input.Name,
		Details:   input.Details,
		BelongsTo: input.BelongsTo,
	}

	query, args := p.buildCreateItemQuery(i)

	// create the item
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&i.ID, &i.CreatedOn)
	if err != nil {
		return nil, errors.Wrap(err, "error executing item creation query")
	}

	return i, nil
}

func (p *Postgres) buildUpdateItemQuery(input *models.Item) (string, []interface{}) {
	query, args, err := p.sqlBuilder.Update(itemsTableName).
		Set("name", input.Name).
		Set("details", input.Details).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{
			"id":         input.ID,
			"belongs_to": input.BelongsTo,
		}).
		Suffix("RETURNING updated_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// UpdateItem updates a particular item. Note that UpdateItem expects the provided input to have a valid ID.
func (p *Postgres) UpdateItem(ctx context.Context, input *models.Item) error {
	query, args := p.buildUpdateItemQuery(input)
	return p.db.QueryRowContext(ctx, query, args...).Scan(&input.UpdatedOn)
}

func (p *Postgres) buildArchiveItemQuery(itemID, userID uint64) (string, []interface{}) {
	query, args, err := p.sqlBuilder.
		Update(itemsTableName).
		Set("updated_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Set("archived_on", squirrel.Expr("extract(epoch FROM NOW())")).
		Where(squirrel.Eq{
			"id":          itemID,
			"belongs_to":  userID,
			"archived_on": nil,
		}).
		Suffix("RETURNING archived_on").
		ToSql()

	logQueryBuildingError(p.logger, err)

	return query, args
}

// DeleteItem deletes an item from the db by its ID
func (p *Postgres) DeleteItem(ctx context.Context, itemID uint64, userID uint64) error {
	query, args := p.buildArchiveItemQuery(itemID, userID)
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}
