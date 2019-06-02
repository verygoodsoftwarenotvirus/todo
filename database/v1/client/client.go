package dbclient

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
)

var _ database.Database = (*Client)(nil)

/*
	NOTE: the original purpose of this client is to allow convenient wrapping of actual query execution.

	I didn't want to neglect one querier implementation while another flourished, so I created this wrapper.
	In reality, a better abstraction is needed, but I need some time to think about what that should look like,
	but it probably will look more like this and less like it did.
*/

// Client is a wrapper around a querier
type Client struct {
	db      *sql.DB
	querier database.Database

	debug  bool
	logger logging.Logger
}

// Migrate is a simple wrapper around the core querier Migrate call
func (c *Client) Migrate(ctx context.Context) error {
	return c.querier.Migrate(ctx)
}

// IsReady is a simple wrapper around the core querier IsReady call
func (c *Client) IsReady(ctx context.Context) (ready bool) {
	return c.querier.IsReady(ctx)
}

// ProvideDatabaseClient provides a querier client
func ProvideDatabaseClient(
	db *sql.DB,
	database database.Database,
	debug bool,
	logger logging.Logger,
) (database.Database, error) {
	c := &Client{
		db:      db,
		querier: database,
		debug:   debug,
		logger:  logger.WithName("db_client"),
	}

	logger.Debug("migrating querier")
	if err := c.querier.Migrate(context.Background()); err != nil {
		return nil, err
	}
	logger.Debug("querier migrated!")

	return c, nil
}
